package utils

import (
	"os"
	"path/filepath"
	"strings"
)

func ValidateFilePath(dir, filename string) (string, error) {
	// Clean the filename to resolve any ".." or unnecessary path elements
	filePath := filepath.Join(dir, filepath.Clean(filename))
	// Ensure the resolved file path starts with the directory
	if !strings.HasPrefix(filePath, dir) {
		return "", os.ErrPermission // Deny access if outside directory
	}

	// Check if the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", os.ErrNotExist // File does not exist
	}

	return filePath, nil
}
