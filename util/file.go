package util

import (
	"io/fs"
	"os"
	"path/filepath"
)

func GetFiles(root, extension string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Check if the file has a .jsx extension
		if !d.IsDir() && filepath.Ext(path) == extension {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

func GetFileContent(root, extension string) (map[string]string, error) {
	files := map[string]string{}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Check if the file has a .jsx extension
		if !d.IsDir() && filepath.Ext(path) == extension {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			// 确保路径是相对于root的
			relPath, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			files[relPath] = string(content)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}
