package internal

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func fileList(dir string, includes []regexp.Regexp, excludes []regexp.Regexp) ([]string, error) {
	defaultExclude := regexp.MustCompile(`\/\.git-ops-update(\.cache)?\.yaml$`)
	files := []string{}
	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		pathRel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		pathRel = "/" + pathRel

		for _, i := range includes {
			if i.Match([]byte(pathRel)) {
				excluded := false
				if defaultExclude.MatchString(pathRel) {
					excluded = true
				}
				for _, e := range excludes {
					if e.MatchString(pathRel) {
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
	return files, err
}

func FileResolvePath(dir string, file string) string {
	if !filepath.IsAbs(file) {
		return filepath.Join(dir, file)
	}
	return file
}

func validateName(name string) bool {
	nameRegex := regexp.MustCompile(`^([a-z0-9\-]+)$`)
	return nameRegex.MatchString(name)
}

func runCallbacks(callbacks []func() error) error {
	for _, cb := range callbacks {
		err := cb()
		if err != nil {
			return err
		}
	}
	return nil
}

func trimRightMultilineString(str string, cutset string) string {
	lines := strings.Split(str, "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], cutset)
	}
	return strings.Join(lines, "\n")
}

func sliceMap[E any, E2 any](slice []E, mapFn func(e E) E2) []E2 {
	result := []E2{}
	for _, e := range slice {
		result = append(result, mapFn(e))
	}
	return result
}

func sliceUnique[E comparable](slice []E) []E {
	values := map[E]bool{}
	result := []E{}
	for _, e := range slice {
		if _, v := values[e]; !v {
			values[e] = true
			result = append(result, e)
		}
	}
	return result
}
