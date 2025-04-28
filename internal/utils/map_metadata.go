package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type MapMetadata struct {
	MapID string `json:"map_id"`
}

func SaveMapMetadata(mapID string, folderPath string) error {
	metadata := MapMetadata{
		MapID: mapID,
	}

	data, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal map metadata: %w", err)
	}

	metadataPath := filepath.Join(folderPath, "golain.json")
	if err := os.WriteFile(metadataPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write map metadata: %w", err)
	}

	return nil
}

func LoadMapMetadata(folderPath string) (*MapMetadata, error) {
	metadataPath := filepath.Join(folderPath, "golain.json")
	
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read map metadata: %w", err)
	}

	var metadata MapMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal map metadata: %w", err)
	}

	return &metadata, nil
} 