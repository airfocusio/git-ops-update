package internal

import (
	"fmt"
	"strings"
)

type Format interface {
	ExtractVersion(str string) (*string, error)
	ReplaceVersion(str string, version string) (*string, error)
}

type PlainFormat struct{}

func (f PlainFormat) ExtractVersion(str string) (*string, error) {
	return &str, nil
}

func (f PlainFormat) ReplaceVersion(str string, version string) (*string, error) {
	return &version, nil
}

type TagFormat struct{}

func (f TagFormat) ExtractVersion(str string) (*string, error) {
	segments := strings.Split(str, ":")
	if len(segments) != 2 {
		return nil, fmt.Errorf("value %s is not a in a valid tag-like format", str)
	}
	return &segments[1], nil
}

func (f TagFormat) ReplaceVersion(str string, version string) (*string, error) {
	segments := strings.Split(str, ":")
	if len(segments) != 2 {
		return nil, fmt.Errorf("value %s is not a in a valid tag-like format", str)
	}
	result := segments[0] + ":" + version
	return &result, nil
}

func getFormat(formatName string) (*Format, error) {
	switch formatName {
	case "":
		return getFormat("plain")
	case "plain":
		format := Format(PlainFormat{})
		return &format, nil
	case "tag":
		format := Format(TagFormat{})
		return &format, nil
	default:
		return nil, fmt.Errorf("unknown format %s", formatName)
	}
}
