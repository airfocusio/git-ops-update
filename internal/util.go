package internal

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
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

func readYaml(bytes []byte, v interface{}) error {
	return yaml.Unmarshal(bytes, v)
}

func writeYaml(v interface{}) ([]byte, error) {
	var bs bytes.Buffer
	enc := yaml.NewEncoder(&bs)
	enc.SetIndent(2)
	err := enc.Encode(v)
	return bs.Bytes(), err
}
