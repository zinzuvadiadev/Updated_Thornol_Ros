package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type ObjectPushPayload struct {
	URL      string                 `json:"url"`
	ObjectID string                 `json:"object_id"`
	Metadata map[string]interface{} `json:"metadata"`
}

// DownloadFile downloads a file from URL and saves it to the specified path asynchronously
func DownloadFile(url, filepath string) chan error {
	errChan := make(chan error, 1)

	go func() {
		defer close(errChan)

		resp, err := http.Get(url)
		if err != nil {
			errChan <- fmt.Errorf("failed to download file: %w", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			errChan <- fmt.Errorf("bad status: %s", resp.Status)
			return
		}

		out, err := os.Create(filepath)
		if err != nil {
			errChan <- fmt.Errorf("failed to create file: %w", err)
			return
		}
		defer out.Close()

		_, err = io.Copy(out, resp.Body)
		if err != nil {
			errChan <- err
			return
		}

		errChan <- nil
	}()

	return errChan
}

// CopyMapFiles downloads and copies map files to the navigation stack folder
func CopyMapFiles(payload *ObjectPushPayload, destFolder string) error {
	// Create destination folder if it doesn't exist
	if err := os.MkdirAll(destFolder, 0755); err != nil {
		return fmt.Errorf("failed to create destination folder: %w", err)
	}
	var destPath string
	if objects, ok := payload.Metadata["objects"].(map[string]interface{}); ok {
		if pgmID, ok := objects["PGM"].(string); ok && pgmID == payload.ObjectID {
			destPath = filepath.Join(destFolder, "map.pgm")
		} else if yamlID, ok := objects["YAML"].(string); ok && yamlID == payload.ObjectID {
			destPath = filepath.Join(destFolder, "golain_map.yaml")
		} else {
			destPath = filepath.Join(destFolder, payload.ObjectID)
		}
	} else {
		destPath = filepath.Join(destFolder, payload.ObjectID)
	}

	// Start the download asynchronously
	errChan := DownloadFile(payload.URL, destPath)

	// Wait for the download to complete
	if err := <-errChan; err != nil {
		return fmt.Errorf("failed to download %s: %w", payload.URL, err)
	}

	return nil
}
