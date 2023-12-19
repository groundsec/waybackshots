package utils

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
)

// IsURL checks if a string is a valid URL.
func IsURL(str string) bool {
	u, err := url.ParseRequestURI(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

// GetDomain extracts the domain from a URL string.
func GetDomain(urlStr string) (string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	// Split the host into parts and return the domain.
	parts := strings.Split(u.Host, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid domain")
	}

	// This simple approach assumes a standard domain structure.
	domain := parts[len(parts)-2] + "." + parts[len(parts)-1]
	return domain, nil
}

// SanitizeFilename removes the protocol and the slashes from a URL string
func SanitizeFilename(urlStr string) string {
	sanitized := strings.Split(urlStr, "://")[1]
	re, _ := regexp.Compile(`/+`) // Matches one or more '/' characters
	return re.ReplaceAllString(sanitized, "-")
}

// CreateFolderIfNotExist creates a folder at the specified path if it does not already exist.
func CreateFolderIfNotExist(folderPath string) error {
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		// Folder does not exist, create it.
		return os.MkdirAll(folderPath, 0755) // 0755 is a common permission setting (read and execute permission for everyone and also write permission for the owner of the file)
	}
	return nil
}
