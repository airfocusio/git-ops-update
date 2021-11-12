package internal

import (
	"os"
	"path/filepath"
)

func fileList(dir string, includes []string, excludes []string) (*[]string, error) {
	temp := []string{}
	for _, exclude := range excludes {
		fs, err := fileGlob(dir, exclude)
		if err != nil {
			return nil, err
		}
		temp = append(temp, *fs...)
	}
	files := []string{}
	for _, include := range includes {
		fs, err := fileGlob(dir, include)
		if err != nil {
			return nil, err
		}
		for _, f := range *fs {
			excluded := false
			for _, f2 := range temp {
				if f == f2 {
					excluded = true
					break
				}
			}
			if !excluded {
				files = append(files, f)
			}
		}
	}
	return &files, nil
}

func fileGlob(dir string, pattern string) (*[]string, error) {
	files := []string{}
	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		_, file := filepath.Split(path)
		matched, err := filepath.Match(pattern, file)
		if err != nil {
			return err
		}
		if matched {
			files = append(files, path)
		}
		return nil
	})
	return &files, err
}
