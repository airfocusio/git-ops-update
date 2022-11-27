package internal

import (
	"fmt"
	"regexp"
	"strings"
)

type Format interface {
	ExtractVersion(str string) (*string, error)
	ReplaceVersion(str string, version string) (*string, error)
}

var _ Format = (*PlainFormat)(nil)

type PlainFormat struct{}

func (f PlainFormat) ExtractVersion(str string) (*string, error) {
	return &str, nil
}

func (f PlainFormat) ReplaceVersion(str string, version string) (*string, error) {
	return &version, nil
}

var _ Format = (*DockerImageFormat)(nil)

type DockerImageFormat struct{}

func (f DockerImageFormat) ExtractVersion(str string) (*string, error) {
	segments := strings.Split(str, ":")
	if len(segments) != 2 {
		return nil, fmt.Errorf("value %s is not a in a valid docker-image format", str)
	}
	return &segments[1], nil
}

func (f DockerImageFormat) ReplaceVersion(str string, version string) (*string, error) {
	segments := strings.Split(str, ":")
	if len(segments) != 2 {
		return nil, fmt.Errorf("value %s is not a in a valid docker-image format", str)
	}
	result := segments[0] + ":" + version
	return &result, nil
}

var _ Format = (*RegexpFormat)(nil)

type RegexpFormat struct {
	Pattern regexp.Regexp
}

func (f RegexpFormat) ExtractVersion(str string) (*string, error) {
	match := f.Pattern.FindStringSubmatch(str)
	if match == nil {
		return nil, fmt.Errorf("value %s is not a in a valid according to regex pattern %s", str, &f.Pattern)
	}
	return &match[f.Pattern.SubexpIndex("version")], nil
}

func (f RegexpFormat) ReplaceVersion(str string, version string) (*string, error) {
	match := f.Pattern.FindAllStringSubmatchIndex(str, -1)
	if match == nil {
		return nil, fmt.Errorf("value %s is not a in a valid according to regex pattern %s", str, &f.Pattern)
	}
	result := str
	names := f.Pattern.SubexpNames()
	delta := 0
	for i := 0; i < len(match); i++ {
		for j := 1; j < len(match[i])/2; j++ {
			i1 := match[i][j*2]
			i2 := match[i][j*2+1]
			if names[j] == "version" && i1 >= 0 && i2 >= 0 {
				result = result[:(i1+delta)] + version + result[(i2+delta):]
				delta = delta + len(version) - i2 + i1
			}
		}
	}
	return &result, nil
}

func getFormat(formatName string) (*Format, error) {
	if formatName == "" {
		return getFormat("plain")
	}

	if formatName == "plain" {
		format := Format(PlainFormat{})
		return &format, nil
	}
	if formatName == "docker-image" {
		format := Format(DockerImageFormat{})
		return &format, nil
	}
	if strings.HasPrefix(formatName, "regexp:") {
		pattern, err := regexp.Compile(strings.TrimPrefix(formatName, "regexp:"))
		if err != nil {
			return nil, err
		}
		if pattern.SubexpIndex("version") < 0 {
			return nil, fmt.Errorf("regexp must contain at least one group with name 'version'")
		}
		format := Format(RegexpFormat{Pattern: *pattern})
		return &format, nil
	}

	return nil, fmt.Errorf("unknown format %s", formatName)
}
