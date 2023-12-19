// Command screenshot is a chromedp example demonstrating how to take a
// screenshot of a specific element and of the entire browser viewport.
package screenshot

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"sync"

	"github.com/chromedp/chromedp"
	"github.com/groundsec/waybackshots/pkg/logger"
	"github.com/groundsec/waybackshots/pkg/utils"
)

type WaybackMachineJSON [][]string

// fullScreenshot takes a screenshot of the entire browser viewport.
//
// Note: chromedp.FullScreenshot overrides the device's emulation settings. Use
// device.Reset to reset the emulation and viewport settings.
func fullScreenshot(url string, quality int, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.FullScreenshot(res, quality),
	}
}

// Worker function to process elements.
func screenshotWorker(id int, url string, timestamps <-chan string, wg *sync.WaitGroup) {
	for timestamp := range timestamps {
		fmt.Printf("Worker %d processing element: %s\n", id, timestamp)
		// create context
		ctx, cancel := chromedp.NewContext(
			context.Background(),
			// chromedp.WithDebugf(log.Printf),
		)

		// capture screenshot of an element
		var buf []byte
		waybackUrl := fmt.Sprintf("http://web.archive.org/web/%s/%s", timestamp, url)
		// capture entire browser viewport, returning png with quality=90
		if err := chromedp.Run(ctx, fullScreenshot(waybackUrl, 90, &buf)); err != nil {
			cancel()
			log.Fatal(err)
		}
		// Get the domain from the URL
		domain, err := utils.GetDomain(url)
		if err != nil {
			logger.Fatal(fmt.Sprintf("Unable to obtain domain from url %s", domain))
		}

		utils.CreateFolderIfNotExist(fmt.Sprintf("waybackshots_%s/%s", domain, timestamp))
		if err := os.WriteFile(fmt.Sprintf("waybackshots_%s/%s/%s.png", domain, timestamp, utils.SanitizeFilename(url)), buf, 0o644); err != nil {
			log.Fatal(err)
		}
	}
	wg.Done()
}

// Fetches and processes data in parallel using worker pool.
func HandleUrl(url string, numWorkers int) {
	if !utils.IsURL(url) {
		logger.Error(fmt.Sprintf("%s is not a URL", url))
		return
	}
	resp, err := http.Get(fmt.Sprintf("http://web.archive.org/cdx/search/cdx?url=%s&output=json&fl=timestamp,digest", url))
	if err != nil {
		logger.Error(fmt.Sprintf("Error fetching data from Wayback Machine: %s\n", err))
		return
	}
	defer resp.Body.Close()

	var jsonResponse WaybackMachineJSON
	err = json.NewDecoder(resp.Body).Decode(&jsonResponse)
	if err != nil {
		logger.Error(fmt.Sprintf("Error decoding JSON from Wayback Machine: %s\n", err))
		return
	}

	digests := []string{}
	timestamps := []string{}
	for i, element := range jsonResponse {
		if i != 0 && !slices.Contains(digests, element[1]) {
			digests = append(digests, element[1])
			timestamps = append(timestamps, element[0])
		}
	}

	var wg sync.WaitGroup
	timestampsChan := make(chan string, len(jsonResponse))

	// Start workers.
	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go screenshotWorker(w, url, timestampsChan, &wg)
	}

	// Distribute work.
	for _, timestamp := range timestamps {
		timestampsChan <- timestamp
	}
	close(timestampsChan)

	// Wait for all workers to finish.
	wg.Wait()

}

func HandleFile(file string, numWorkers int) {
	f, err := os.Open(file)
	if err != nil {
		logger.Fatal("Error opening file:", err)
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		HandleUrl(scanner.Text(), numWorkers)
	}
	if err := scanner.Err(); err != nil {
		logger.Fatal("Error reading file:", err)
	}
}
