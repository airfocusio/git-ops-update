package main

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"

	"github.com/blang/semver/v4"
)

// LexicographicExtractStrategyConfig ...
type LexicographicExtractStrategyConfig struct{}

// NumericExtractStrategyConfig ...
type NumericExtractStrategyConfig struct{}

// SemverExtractStrategyConfig ...
type SemverExtractStrategyConfig struct {
	PinMajor         bool `json:"pinMajor"`
	PinMinor         bool `json:"pinMinor"`
	PinPatch         bool `json:"pinPatch"`
	AllowPrereleases bool `json:"allowPrereleases"`
}

// ExtractConfig ...
type ExtractConfig struct {
	Value         string                              `json:"value"`
	Lexicographic *LexicographicExtractStrategyConfig `json:"lexicographic"`
	Numeric       *NumericExtractStrategyConfig       `json:"numeric"`
	Semver        *SemverExtractStrategyConfig        `json:"semver"`
}

// PolicyConfig ...
type PolicyConfig struct {
	Name     string          `json:"name"`
	Pattern  string          `json:"pattern"`
	Extracts []ExtractConfig `json:"extracts"`
}

// Policy ...
type Policy struct {
	Name     string
	Pattern  *regexp.Regexp
	Extracts []Extract
}

// Extract
type Extract struct {
	Value    string
	Strategy ExtractStrategy
}

// ExtractStrategy ...
type ExtractStrategy interface {
	IsCompatible(v1 string, v2 string) bool
	Compare(v1 string, v2 string) int
}

// LexicographicExtractStrategy ...
type LexicographicExtractStrategy struct {
	Config LexicographicExtractStrategyConfig
}

// NumericExtractStrategy ...
type NumericExtractStrategy struct {
	Config NumericExtractStrategyConfig
}

// SemverExtractStrategy ...
type SemverExtractStrategy struct {
	Config SemverExtractStrategyConfig
}

var extractPattern = regexp.MustCompile(`<([a-zA-Z0-9\-]+)>`)

// Parse ...
func (p Policy) Parse(version string) (*[]string, error) {
	segments := map[string]string{}
	if p.Pattern != nil {
		match := p.Pattern.FindStringSubmatch(version)
		if match == nil {
			return &[]string{}, fmt.Errorf("version %s does not match pattern %v", version, p.Pattern)
		}
		names := p.Pattern.SubexpNames()
		for i, s := range match {
			segments[names[i]] = s
		}
	}

	result := []string{}
	for _, e := range p.Extracts {
		value := extractPattern.ReplaceAllStringFunc(e.Value, func(raw string) string {
			key := raw[1 : len(raw)-1]
			value := segments[key]
			return value
		})
		result = append(result, value)
	}

	return &result, nil
}

type versionParsed struct {
	Version string
	Parsed  []string
}
type versionParsedList struct {
	Extracts []Extract
	Items    []versionParsed
}

func (l versionParsedList) Len() int {
	return len(l.Items)
}
func (l versionParsedList) Swap(i, j int) {
	l.Items[i], l.Items[j] = l.Items[j], l.Items[i]
}
func (l versionParsedList) Less(i, j int) bool {
	a := l.Items[i]
	b := l.Items[j]
	for i, e := range l.Extracts {
		cmp := e.Strategy.Compare(a.Parsed[i], b.Parsed[i])
		if cmp > 0 {
			return true
		}
		if cmp < 0 {
			return false
		}
	}
	return false
}

// FilterAndSort ...
func (p Policy) FilterAndSort(currentVersion string, availableVersions []string) (*[]string, error) {
	currentVersionParsed, err := p.Parse(currentVersion)
	if err != nil {
		return nil, err
	}

	temp1 := []versionParsed{}
	for _, version := range availableVersions {
		parsed, err := p.Parse(version)
		if err == nil {
			temp1 = append(temp1, versionParsed{
				Version: version,
				Parsed:  *parsed,
			})
		}
	}
	temp2 := versionParsedList{
		Items:    temp1,
		Extracts: p.Extracts,
	}
	sort.Sort(temp2)

	result := []string{}
	for _, version := range temp2.Items {
		isCompatible := true
		for i, parsed := range version.Parsed {
			if !temp2.Extracts[i].Strategy.IsCompatible((*currentVersionParsed)[i], parsed) {
				isCompatible = false
			}
		}
		if isCompatible {
			result = append(result, version.Version)
		}
	}

	return &result, nil
}

func (p Policy) FindNext(currentVersion string, availableVersions []string) (*string, error) {
	allVersions := append(availableVersions, currentVersion)
	allFilteredSortedVersions, err := p.FilterAndSort(currentVersion, allVersions)
	if err != nil {
		return nil, err
	}
	if len(*allFilteredSortedVersions) > 0 {
		return &(*allFilteredSortedVersions)[0], nil
	}
	return &currentVersion, nil
}

// Compare ...
func (str LexicographicExtractStrategy) Compare(v1 string, v2 string) int {
	if v1 == v2 {
		return 0
	}
	if v1 > v2 {
		return 1
	}
	if v1 < v2 {
		return -1
	}
	return 0
}

// IsCompatible ...
func (str LexicographicExtractStrategy) IsCompatible(v1 string, v2 string) bool {
	return true
}

// Compare ...
func (str NumericExtractStrategy) Compare(v1 string, v2 string) int {
	if v1 == v2 {
		return 0
	}
	v1i, v1e := strconv.Atoi(v1)
	v2i, v2e := strconv.Atoi(v2)
	if v1e != nil {
		return -1
	}
	if v2e != nil {
		return 1
	}
	if v1i > v2i {
		return 1
	}
	if v1i < v2i {
		return -1
	}
	return 0
}

// IsCompatible ...
func (str NumericExtractStrategy) IsCompatible(v1 string, v2 string) bool {
	return true
}

// Compare ...
func (str SemverExtractStrategy) Compare(v1 string, v2 string) int {
	if v1 == v2 {
		return 0
	}
	v1sv, _ := semver.Make(v1)
	v2sv, _ := semver.Make(v2)
	return v1sv.Compare(v2sv)
}

// IsCompatible ...
func (str SemverExtractStrategy) IsCompatible(v1 string, v2 string) bool {
	v1sv, _ := semver.Make(v1)
	v2sv, _ := semver.Make(v2)
	if str.Config.PinMajor && v1sv.Major != v2sv.Major {
		return false
	}
	if str.Config.PinMinor && (v1sv.Major != v2sv.Major || v1sv.Minor != v2sv.Minor) {
		return false
	}
	if str.Config.PinPatch && (v1sv.Major != v2sv.Major || v1sv.Minor != v2sv.Minor || v1sv.Patch != v2sv.Patch) {
		return false
	}
	return true
}
