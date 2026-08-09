package main

import (
	"slices"
	"encoding/json"
	"os"
	"path/filepath"
)

type Manifest struct {
    Version int      `json:"version"`
    Files   []string `json:"files"`
}

func loadManifest(repo *Repository) (*Manifest, error) {
	fullPath := filepath.Join(repo.Root, ".carry.json")
	data, err := os.ReadFile(fullPath)
	
	if err != nil {
        if os.IsNotExist(err) {
            return &Manifest{
                Version: 1,
                Files:   []string{},
            }, nil
        }

        return nil, err
    }

	var manifest Manifest

	err = json.Unmarshal(data, &manifest)
	if err != nil {
		return nil, err
	}
	return &manifest, nil
	

}

func saveManifest(repo *Repository, manifest *Manifest) error {
	jsonData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	
	fullPath := filepath.Join(repo.Root, ".carry.json")
	err = os.WriteFile(fullPath, jsonData, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (m *Manifest) Add(path string) bool {
	if slices.Contains(m.Files, path) {
			return false
		}
	m.Files = append(m.Files, path)
	return true
}