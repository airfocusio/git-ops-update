package internal

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
)

func fileList(dir string, includes []regexp.Regexp, excludes []regexp.Regexp) (*[]string, error) {
	files := []string{}
	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		pathRel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		for _, i := range includes {
			if i.Match([]byte(pathRel)) {
				excluded := false
				for _, e := range excludes {
					if e.Match([]byte(pathRel)) {
						excluded = true
					}
				}
				if !excluded {
					files = append(files, path)
					return nil
				}
			}
		}
		return nil
	})
	return &files, err
}

func fileResolvePath(dir string, file string) string {
	if !filepath.IsAbs(file) {
		return filepath.Join(dir, file)
	}
	return file
}

func fileReadYaml(file string, v interface{}) error {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	err = readYaml(bytes, v)
	if err != nil {
		return err
	}
	return nil
}

func fileWriteYaml(file string, v interface{}) error {
	bytes, err := writeYaml(v)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(file, bytes, 0644)
	if err != nil {
		return err
	}
	return nil
}
