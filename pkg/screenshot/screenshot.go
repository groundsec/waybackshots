// Command screenshot is a chromedp example demonstrating how to take a
// screenshot of a specific element and of the entire browser viewport.
package screenshot

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"slices"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/groundsec/waybackshots/pkg/logger"
	"github.com/groundsec/waybackshots/pkg/utils"
)

type WaybackMachineJSON [][]string
type URLRecordInfo struct {
	Timestamp string
	Digest    string
	Original  string
}

var defaultTimeout = 2
var errorTimeout = 30
var connectionErrors = 0

func extractDigests(recordInfo []URLRecordInfo) []string {
	digests := []string{}
	for _, record := range recordInfo {
		if !slices.Contains(digests, record.Digest) {
			digests = append(digests, record.Digest)
		}
	}
	return digests
}

// fullScreenshot takes a screenshot of the entire browser viewport.
//
// Note: chromedp.FullScreenshot overrides the device's emulation settings. Use
// device.Reset to reset the emulation and viewport settings.
func fullScreenshot(urlStr string, quality int, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlStr),
		chromedp.FullScreenshot(res, quality),
	}
}

// Worker function to process elements.
func screenshotWorker(url string, record URLRecordInfo, userAgent string) bool {
	waybackUrl := fmt.Sprintf("https://web.archive.org/web/%sid_/%s", record.Timestamp, url)
	logger.Info(fmt.Sprintf("Processing element: %s\n", waybackUrl))
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent(userAgent),
		//chromedp.ProxyServer("http://localhost:8080"),
		chromedp.IgnoreCertErrors,
	)
	ctx, cancel := chromedp.NewExecAllocator(
		context.Background(),
		opts...,
	)
	defer cancel()
	// Create chromedp context
	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	// capture screenshot of an element
	var buf []byte
	// capture entire browser viewport, returning png with quality=90
	if err := chromedp.Run(ctx, fullScreenshot(waybackUrl, 90, &buf)); err != nil {
		logger.Debug(err)
		return false
	}
	// Get the domain from the URL
	domain, err := utils.GetDomain(url)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Unable to obtain domain from url %s", domain))
	}

	utils.CreateFolderIfNotExist(fmt.Sprintf("waybackshots_%s/%s", domain, record.Timestamp))
	if err := os.WriteFile(fmt.Sprintf("waybackshots_%s/%s/%s_%s.png", domain, record.Timestamp, utils.SanitizeFilename(url), record.Digest), buf, 0o644); err != nil {
		log.Fatal(err)
	}
	return true
}

// Fetches and processes data in parallel using worker pool.
func HandleUrl(urlString string) {
	if !utils.IsURL(urlString) {
		logger.Error(fmt.Sprintf("%s is not a URL", urlString))
		return
	}
	userAgent := utils.UserAgents[rand.Intn(len(utils.UserAgents))]

	/*
		proxyURLParsed, err := url.Parse("http://localhost:8081")
		if err != nil {
			fmt.Errorf("invalid proxy URL: %w", err)
			return
		}
	*/

	transport := &http.Transport{
		//Proxy: http.ProxyURL(proxyURLParsed),
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	client := &http.Client{
		Transport: transport,
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("https://web.archive.org/cdx/search/cdx?url=%s&output=json&fl=timestamp,digest", urlString), nil)
	if err != nil {
		logger.Error("Unable to create GET request")
		return
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Error fetching data from Wayback Machine, retry in a little bit")
		return
	}
	defer resp.Body.Close()

	var jsonResponse WaybackMachineJSON
	err = json.NewDecoder(resp.Body).Decode(&jsonResponse)
	if err != nil {
		logger.Error(fmt.Sprintf("Error decoding JSON from Wayback Machine: %s\n", err))
		return
	}

	recordData := []URLRecordInfo{}
	for i, element := range jsonResponse {
		jsonTimestamp := element[0]
		jsonDigest := element[1]
		if i != 0 && !slices.Contains(extractDigests(recordData), jsonDigest) {
			recordData = append(recordData, URLRecordInfo{Timestamp: jsonTimestamp, Digest: jsonDigest})
		}
	}

	for _, record := range recordData {
		for {
			screenshotCompleted := screenshotWorker(urlString, record, userAgent)
			if screenshotCompleted {
				connectionErrors = 0
				time.Sleep(time.Duration(defaultTimeout) * time.Second)
				break
			} else {
				connectionErrors += 1
				logger.Error(fmt.Sprintf("Connection error on URL '%s' with digest %s, waiting for %d seconds", urlString, record.Digest, connectionErrors*errorTimeout))
				time.Sleep(time.Duration(connectionErrors*errorTimeout) * time.Second)
			}
		}
	}
	fmt.Printf("Wayback Machine screenshots completed for URL %s\n", urlString)
}

func handleDomain(domain string, urlStrings []string) {
	userAgent := utils.UserAgents[rand.Intn(len(utils.UserAgents))]

	/*
		proxyURLParsed, err := url.Parse("http://localhost:8081")
		if err != nil {
			fmt.Errorf("invalid proxy URL: %w", err)
			return
		}
	*/

	transport := &http.Transport{
		//Proxy: http.ProxyURL(proxyURLParsed),
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	client := &http.Client{
		Transport: transport,
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("https://web.archive.org/cdx/search/cdx?url=%s/*&output=json&fl=original,timestamp,digest", domain), nil)
	if err != nil {
		logger.Error("Unable to create GET request")
		return
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Error fetching data from Wayback Machine, retry in a little bit")
		return
	}
	defer resp.Body.Close()

	var jsonResponse WaybackMachineJSON
	err = json.NewDecoder(resp.Body).Decode(&jsonResponse)
	if err != nil {
		logger.Error(fmt.Sprintf("Error decoding JSON from Wayback Machine: %s\n", err))
		return
	}

	for _, urlString := range urlStrings {
		recordData := []URLRecordInfo{}
		for i, element := range jsonResponse {
			jsonOriginal := element[0]
			jsonTimestamp := element[1]
			jsonDigest := element[2]
			if i != 0 && !slices.Contains(extractDigests(recordData), jsonDigest) && urlString == jsonOriginal {
				recordData = append(recordData, URLRecordInfo{Timestamp: jsonTimestamp, Digest: jsonTimestamp})
			}
		}
		for _, record := range recordData {
			screenshotCompleted := screenshotWorker(urlString, record, userAgent)
			if screenshotCompleted {
				connectionErrors = 0
				time.Sleep(time.Duration(defaultTimeout) * time.Second)
				break
			} else {
				connectionErrors += 1
				logger.Error(fmt.Sprintf("Connection error on URL '%s' with digest %s, waiting for %d seconds", urlString, record.Digest, connectionErrors*errorTimeout))
				time.Sleep(time.Duration(connectionErrors*errorTimeout) * time.Second)
			}
		}
	}
}

func HandleFile(file string) {
	f, err := os.Open(file)
	if err != nil {
		logger.Fatal("Error opening file:", err)
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	domainURLs := make(map[string][]string)

	for scanner.Scan() {
		lineStr := scanner.Text()
		if !utils.IsURL(lineStr) {
			logger.Error(fmt.Sprintf("Line '%s' is not a valid URL", lineStr))
			continue
		}
		domain, err := utils.GetDomain(lineStr)
		if err != nil {
			continue
		}

		if _, exists := domainURLs[domain]; !exists {
			domainURLs[domain] = []string{}
		}

		domainURLs[domain] = append(domainURLs[domain], lineStr)
	}
	for domain, values := range domainURLs {
		handleDomain(domain, values)
	}
	if err := scanner.Err(); err != nil {
		logger.Fatal("Error reading file:", err)
	}
	fmt.Printf("Wayback Machine screenshots completed for file %s\n", file)
}
