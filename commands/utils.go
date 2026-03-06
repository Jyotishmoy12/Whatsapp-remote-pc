package commands

import (
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/vova616/screenshot"
)

func CaptureScreen() (string, error) {
	img, err := screenshot.CaptureScreen()

	if err != nil {
		return "", err
	}
	fileName := "current_screen.png"

	f, err := os.Create(fileName)
	if err != nil {
		return "", err
	}
	defer f.Close()

	err = png.Encode(f, img)
	if err != nil {
		return "", err
	}
	return fileName, nil
}

func FindFile(root, targetName string) ([]string, error) {
	var matches []string

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // Skip folders we can't access
		}

		// If the filename matches (case-insensitive for Windows)
		if strings.EqualFold(d.Name(), targetName) {
			matches = append(matches, path)
		}
		return nil
	})

	return matches, err
}
