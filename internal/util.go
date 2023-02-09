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

func SliceMap[E any, E2 any](slice []E, mapFn func(e E) E2) []E2 {
	result := []E2{}
	for _, e := range slice {
		result = append(result, mapFn(e))
	}
	return result
}

func SliceFilter[E any](slice []E, filterFn func(e E) bool) []E {
	result := []E{}
	for _, e := range slice {
		if filterFn(e) {
			result = append(result, e)
		}
	}
	return result
}

func SliceFlatMap[E any, E2 any](slice []E, mapFn func(e E) []E2) []E2 {
	result := []E2{}
	for _, e := range slice {
		result = append(result, mapFn(e)...)
	}
	return result
}

func SliceUnique[E comparable](slice []E) []E {
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

func SliceGroupBy[E any](slice []E, keyFn func(e E) string) map[string][]E {
	result := map[string][]E{}
	for _, e := range slice {
		key := keyFn(e)
		if _, ok := result[key]; !ok {
			result[key] = []E{e}
		} else {
			result[key] = append(result[key], e)
		}
	}
	return result
}

func MapMap[K comparable, V any, E any](slice map[K]V, mapFn func(v V, k K) E) []E {
	result := []E{}
	for k, v := range slice {
		result = append(result, mapFn(v, k))
	}
	return result
}
