package internal

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/blang/semver/v4"
)

type Policy struct {
	Pattern  *regexp.Regexp
	Extracts []Extract
}

// Extract
type Extract struct {
	Value    string
	Strategy ExtractStrategy
}

type ExtractStrategy interface {
	IsValid(v string) bool
	IsCompatible(v1 string, v2 string) bool
	Compare(v1 string, v2 string) int
}

type LexicographicExtractStrategy struct {
	Pin bool
}

type NumericExtractStrategy struct {
	Pin bool
}

type SemverExtractStrategy struct {
	PinMajor         bool
	PinMinor         bool
	PinPatch         bool
	AllowPrereleases bool
}

var extractPattern = regexp.MustCompile(`<([a-zA-Z0-9\-]+)>`)

func (p Policy) Parse(version string, prefix string, suffix string) (*[]string, error) {
	unpackedVersion := version
	if !strings.HasPrefix(unpackedVersion, prefix) {
		return nil, nil
	}
	unpackedVersion = strings.TrimPrefix(unpackedVersion, prefix)
	if !strings.HasSuffix(version, suffix) {
		return nil, nil
	}
	unpackedVersion = strings.TrimSuffix(unpackedVersion, suffix)
	segments := map[string]string{}
	if p.Pattern != nil {
		match := p.Pattern.FindStringSubmatch(unpackedVersion)
		if match == nil {
			return &[]string{}, fmt.Errorf("version %s does not match pattern %v with prefix \"%s\" and suffix \"%s\"", version, p.Pattern, prefix, suffix)
		}
		names := p.Pattern.SubexpNames()
		for i, s := range match {
			segments[names[i]] = s
		}
	}

	result := []string{}
	for _, e := range p.Extracts {
		value := unpackedVersion
		if e.Value != "" {
			value = extractPattern.ReplaceAllStringFunc(e.Value, func(raw string) string {
				key := raw[1 : len(raw)-1]
				value := segments[key]
				return value
			})
		}
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
	cmp2 := strings.Compare(a.Version, b.Version)
	if cmp2 > 0 {
		return true
	}
	if cmp2 < 0 {
		return false
	}
	return false
}

func (p Policy) FilterAndSort(currentVersion string, availableVersions []string, prefix string, suffix string) (*[]string, error) {
	currentVersionParsed, err := p.Parse(currentVersion, prefix, suffix)
	if err != nil {
		return nil, err
	}
	if currentVersionParsed == nil {
		return nil, fmt.Errorf("version %s does not match pattern %v with prefix \"%s\" and suffix \"%s\"", currentVersion, p.Pattern, prefix, suffix)
	}

	temp1 := []versionParsed{}
	for _, version := range availableVersions {
		parsed, err := p.Parse(version, prefix, suffix)
		if parsed != nil && err == nil {
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
			if !temp2.Extracts[i].Strategy.IsValid((*currentVersionParsed)[i]) {
				return nil, fmt.Errorf("version %s has extraction %s which is invalid for selected strategy", currentVersion, (*currentVersionParsed)[i])
			}
			if !temp2.Extracts[i].Strategy.IsValid(parsed) {
				isCompatible = false
			}
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

func (p Policy) FindNext(currentVersion string, availableVersions []string, prefix string, suffix string) (*string, error) {
	allVersions := append(availableVersions, currentVersion)
	allFilteredSortedVersions, err := p.FilterAndSort(currentVersion, allVersions, prefix, suffix)
	if err != nil {
		return nil, err
	}
	if len(*allFilteredSortedVersions) > 0 {
		return &(*allFilteredSortedVersions)[0], nil
	}
	return &currentVersion, nil
}

func (str LexicographicExtractStrategy) IsValid(v string) bool {
	return true
}

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

func (str LexicographicExtractStrategy) IsCompatible(v1 string, v2 string) bool {
	return !str.Pin || v1 == v2
}

func (str NumericExtractStrategy) IsValid(v string) bool {
	vi, ve := strconv.Atoi(v)
	return ve == nil && vi >= 0
}

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

func (str NumericExtractStrategy) IsCompatible(v1 string, v2 string) bool {
	if !str.IsValid(v1) || !str.IsValid(v2) {
		return false
	}
	return !str.Pin || v1 == v2
}

func (str SemverExtractStrategy) IsValid(v string) bool {
	_, err := semver.Make(v)
	return err == nil
}

func (str SemverExtractStrategy) Compare(v1 string, v2 string) int {
	if v1 == v2 {
		return 0
	}
	v1sv, err1 := semver.Make(v1)
	v2sv, err2 := semver.Make(v2)
	if err1 != nil || err2 != nil {
		return 0
	}
	return v1sv.Compare(v2sv)
}

func (str SemverExtractStrategy) IsCompatible(v1 string, v2 string) bool {
	v1sv, err1 := semver.Make(v1)
	v2sv, err2 := semver.Make(v2)
	if err1 != nil || err2 != nil {
		return false
	}
	if str.PinMajor && v1sv.Major != v2sv.Major {
		return false
	}
	if str.PinMinor && (v1sv.Major != v2sv.Major || v1sv.Minor != v2sv.Minor) {
		return false
	}
	if str.PinPatch && (v1sv.Major != v2sv.Major || v1sv.Minor != v2sv.Minor || v1sv.Patch != v2sv.Patch) {
		return false
	}
	if !str.AllowPrereleases && len(v2sv.Pre) > 0 {
		return false
	}
	return true
}
