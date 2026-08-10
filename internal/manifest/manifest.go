package manifest

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"

	"github.com/gian5204/carry/internal/repo"
)

type Manifest struct {
	Version int      `json:"version"`
	Files   []string `json:"files"`
}

// loads the Carry manifest for a repository
func Load(repository *repo.Repository) (*Manifest, error) {
	fullPath := filepath.Join(repository.Root, ".carry.json")
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

	var currentManifest Manifest

	err = json.Unmarshal(data, &currentManifest)
	if err != nil {
		return nil, err
	}
	return &currentManifest, nil
}

// saves the Carry manifest for a repository
func Save(repository *repo.Repository, currentManifest *Manifest) error {
	jsonData, err := json.MarshalIndent(currentManifest, "", "  ")
	if err != nil {
		return err
	}

	fullPath := filepath.Join(repository.Root, ".carry.json")
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

func (m *Manifest) Remove(path string) bool {
	for i, file := range m.Files {
		if file == path {
			m.Files = append(m.Files[:i], m.Files[i+1:]...)
			return true
		}
	}
	return false
}
