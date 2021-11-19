package internal

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPolicyParse(t *testing.T) {
	p1 := Policy{}
	actual, err := p1.Parse("1", "", "")
	if assert.NoError(t, err) {
		assert.Equal(t, &[]string{}, actual)
	}

	p2 := Policy{Extracts: []Extract{{Strategy: NumericExtractStrategy{}}}}
	actual, err = p2.Parse("1", "", "")
	if assert.NoError(t, err) {
		assert.Equal(t, &[]string{"1"}, actual)
	}
	actual, err = p2.Parse("2", "", "")
	if assert.NoError(t, err) {
		assert.Equal(t, &[]string{"2"}, actual)
	}

	p3 := Policy{
		Pattern: regexp.MustCompile(`^(?P<major>\d+)\.(?P<minor>\d+)$`),
		Extracts: []Extract{
			{
				Value:    "<major>",
				Strategy: NumericExtractStrategy{},
			},
			{
				Value:    "<minor>",
				Strategy: NumericExtractStrategy{},
			},
			{
				Value:    "<major><minor>",
				Strategy: NumericExtractStrategy{},
			},
		},
	}
	actual, err = p3.Parse("1.2", "", "")
	if assert.NoError(t, err) {
		assert.Equal(t, &[]string{"1", "2", "12"}, actual)
	}
	actual, err = p3.Parse("v1.2-alpine", "v", "-alpine")
	if assert.NoError(t, err) {
		assert.Equal(t, &[]string{"1", "2", "12"}, actual)
	}
	actual, err = p3.Parse("v1.2", "v", "-alpine")
	if assert.NoError(t, err) {
		assert.Nil(t, actual)
	}
	actual, err = p3.Parse("1.2-alpine", "v", "-alpine")
	if assert.NoError(t, err) {
		assert.Nil(t, actual)
	}
}

func TestPolicyFilterAndSort(t *testing.T) {
	p1 := Policy{}
	actual, err := p1.FilterAndSort("1", []string{"1", "2", "3"}, "", "")
	if assert.NoError(t, err) {
		assert.Equal(t, []string{"1", "2", "3"}, *actual)
	}

	p2 := Policy{
		Pattern: regexp.MustCompile(`^(?P<major>\d+)\.(?P<minor>\d+)$`),
		Extracts: []Extract{
			{
				Value:    "<major>",
				Strategy: NumericExtractStrategy{},
			},
			{
				Value:    "<minor>",
				Strategy: NumericExtractStrategy{},
			},
		},
	}
	actual, err = p2.FilterAndSort("1.0", strings.Split("18.04 18.10 19.04 19.10 20.04 20.10 21.04 21.10 22.04", " "), "", "")
	if assert.NoError(t, err) {
		assert.Equal(t, strings.Split("22.04 21.10 21.04 20.10 20.04 19.10 19.04 18.10 18.04", " "), *actual)
	}
	actual, err = p2.FilterAndSort("v1.0-ubuntu", strings.Split("17.10 v18.04-ubuntu v18.10-ubuntu v19.04-ubuntu v19.10-ubuntu v20.04-ubuntu v20.10-ubuntu v21.04-ubuntu v21.10-ubuntu v22.04-ubuntu", " "), "v", "-ubuntu")
	if assert.NoError(t, err) {
		assert.Equal(t, strings.Split("v22.04-ubuntu v21.10-ubuntu v21.04-ubuntu v20.10-ubuntu v20.04-ubuntu v19.10-ubuntu v19.04-ubuntu v18.10-ubuntu v18.04-ubuntu", " "), *actual)
	}
	_, err = p2.FilterAndSort("v1.0", strings.Split("17.10 v18.04-ubuntu v18.10-ubuntu v19.04-ubuntu v19.10-ubuntu v20.04-ubuntu v20.10-ubuntu v21.04-ubuntu v21.10-ubuntu v22.04-ubuntu", " "), "v", "-ubuntu")
	assert.Error(t, err)
	_, err = p2.FilterAndSort("1.0-ubuntu", strings.Split("17.10 v18.04-ubuntu v18.10-ubuntu v19.04-ubuntu v19.10-ubuntu v20.04-ubuntu v20.10-ubuntu v21.04-ubuntu v21.10-ubuntu v22.04-ubuntu", " "), "v", "-ubuntu")
	assert.Error(t, err)

	p3 := Policy{
		Pattern: regexp.MustCompile(`^(?P<number>\d+)(?P<rest>.*)?$`),
		Extracts: []Extract{
			{
				Value:    "<number>",
				Strategy: NumericExtractStrategy{},
			},
		},
	}
	actual, err = p3.FilterAndSort("1", strings.Split("1-ab 1 2-ab 2-b 2-a 2-ab 2 1-a 1-b 1-ab", " "), "", "")
	if assert.NoError(t, err) {
		assert.Equal(t, strings.Split("2-b 2-ab 2-ab 2-a 2 1-b 1-ab 1-ab 1-a 1", " "), *actual)
	}
}

func TestPolicyFindNext(t *testing.T) {
	p1 := Policy{}
	actual, err := p1.FindNext("1", []string{"1", "2", "3"}, "", "")
	if assert.NoError(t, err) {
		assert.Equal(t, "1", *actual)
	}

	p2 := Policy{
		Pattern: regexp.MustCompile(`^(?P<major>\d+)\.(?P<minor>\d+)$`),
		Extracts: []Extract{
			{
				Value:    "<major>",
				Strategy: NumericExtractStrategy{},
			},
			{
				Value:    "<minor>",
				Strategy: NumericExtractStrategy{},
			},
		},
	}
	actual, err = p2.FindNext("1.0", []string{"1.0", "1.1", "2.0", "2.0-pre"}, "", "")
	if assert.NoError(t, err) {
		assert.Equal(t, "2.0", *actual)
	}
	actual, err = p2.FindNext("1.1", []string{"1.0", "1.1", "2.0", "2.0-pre"}, "", "")
	if assert.NoError(t, err) {
		assert.Equal(t, "2.0", *actual)
	}
	actual, err = p2.FindNext("2.0", []string{"1.0", "1.1", "2.0", "2.0-pre"}, "", "")
	if assert.NoError(t, err) {
		assert.Equal(t, "2.0", *actual)
	}
	actual, err = p2.FindNext("3.0", []string{"1.0", "1.1", "2.0", "2.0-pre"}, "", "")
	if assert.NoError(t, err) {
		assert.Equal(t, "3.0", *actual)
	}

	p3 := Policy{
		Pattern: regexp.MustCompile(`^(?P<all>.*)$`),
		Extracts: []Extract{
			{
				Value:    "<all>",
				Strategy: SemverExtractStrategy{},
			},
		},
	}
	actual, err = p3.FindNext("1.0.0", []string{"1.0.0", "1.1.0", "2.0.0", "2.0.0-pre"}, "", "")
	if assert.NoError(t, err) {
		assert.Equal(t, "2.0.0", *actual)
	}
	actual, err = p3.FindNext("1.1.0", []string{"1.0.0", "1.1.0", "2.0.0", "2.0.0-pre"}, "", "")
	if assert.NoError(t, err) {
		assert.Equal(t, "2.0.0", *actual)
	}
	actual, err = p3.FindNext("2.0.0", []string{"1.0.0", "1.1.0", "2.0.0", "2.0.0-pre"}, "", "")
	if assert.NoError(t, err) {
		assert.Equal(t, "2.0.0", *actual)
	}
	actual, err = p3.FindNext("3.0.0", []string{"1.0.0", "1.1.0", "2.0.0", "2.0.0-pre"}, "", "")
	if assert.NoError(t, err) {
		assert.Equal(t, "3.0.0", *actual)
	}
}

func TestLexicographicSortStrategyCompare(t *testing.T) {
	str := LexicographicExtractStrategy{}

	assert.Equal(t, 0, str.Compare("1", "1"))
	assert.Equal(t, -1, str.Compare("1", "2"))
	assert.Equal(t, 1, str.Compare("2", "1"))
	assert.Equal(t, -1, str.Compare("1", "10"))
	assert.Equal(t, 1, str.Compare("2", "10"))
	assert.Equal(t, 1, str.Compare("10", "1"))
	assert.Equal(t, -1, str.Compare("10", "2"))
}

func TestLexicographicSortStrategyIsCompatible(t *testing.T) {
	assert.Equal(t, true, LexicographicExtractStrategy{}.IsCompatible("1", "1"))
	assert.Equal(t, true, LexicographicExtractStrategy{}.IsCompatible("1", "2"))
	assert.Equal(t, true, LexicographicExtractStrategy{Pin: true}.IsCompatible("1", "1"))
	assert.Equal(t, false, LexicographicExtractStrategy{Pin: true}.IsCompatible("1", "2"))
}

func TestNumericSortStrategyCompare(t *testing.T) {
	str := NumericExtractStrategy{}

	assert.Equal(t, 0, str.Compare("", ""))
	assert.Equal(t, -1, str.Compare("", "1"))
	assert.Equal(t, 1, str.Compare("1", ""))
	assert.Equal(t, 0, str.Compare("1", "1"))
	assert.Equal(t, -1, str.Compare("1", "2"))
	assert.Equal(t, 1, str.Compare("2", "1"))
	assert.Equal(t, -1, str.Compare("1", "10"))
	assert.Equal(t, -1, str.Compare("2", "10"))
	assert.Equal(t, 1, str.Compare("10", "1"))
	assert.Equal(t, 1, str.Compare("10", "2"))
}

func TestNumericSortStrategyIsCompatible(t *testing.T) {
	assert.Equal(t, true, NumericExtractStrategy{}.IsCompatible("1", "1"))
	assert.Equal(t, true, NumericExtractStrategy{}.IsCompatible("1", "2"))
	assert.Equal(t, true, NumericExtractStrategy{Pin: true}.IsCompatible("1", "1"))
	assert.Equal(t, false, NumericExtractStrategy{Pin: true}.IsCompatible("1", "2"))
}

func TestSemverSortStrategyCompare(t *testing.T) {
	str := SemverExtractStrategy{}

	assert.Equal(t, 0, str.Compare("1.0.0", "1.0.0"))
	assert.Equal(t, -1, str.Compare("1.0.0", "1.0.1"))
	assert.Equal(t, 1, str.Compare("1.0.1", "1.0.0"))
	assert.Equal(t, -1, str.Compare("1.0.1", "2.0.0"))
	assert.Equal(t, 1, str.Compare("2.0.0", "1.0.1"))
	assert.Equal(t, -1, str.Compare("1.0.0", "10.0.0"))
	assert.Equal(t, -1, str.Compare("2.0.0", "10.0.0"))
	assert.Equal(t, 1, str.Compare("10.0.0", "1.0.0"))
	assert.Equal(t, 1, str.Compare("10.0.0", "2.0.0"))
	assert.Equal(t, 0, str.Compare("1.0.0+1", "1.0.0+2"))
	assert.Equal(t, 0, str.Compare("1.0.0+2", "1.0.0+1"))
	assert.Equal(t, -1, str.Compare("1.0.0-dev.2", "1.0.0-dev.10"))
	assert.Equal(t, 1, str.Compare("1.0.0-dev.10", "1.0.0-dev.2"))
}

func TestSemverSortStrategyIsCompatible(t *testing.T) {
	assert.Equal(t, true, SemverExtractStrategy{}.IsCompatible("1.0.0", "1.0.0"))

	assert.Equal(t, true, SemverExtractStrategy{PinMajor: true}.IsCompatible("1.0.0", "1.0.0"))
	assert.Equal(t, true, SemverExtractStrategy{PinMajor: true}.IsCompatible("1.0.0", "1.0.1"))
	assert.Equal(t, true, SemverExtractStrategy{PinMajor: true}.IsCompatible("1.0.0", "1.1.0"))
	assert.Equal(t, false, SemverExtractStrategy{PinMajor: true}.IsCompatible("1.0.0", "2.0.0"))

	assert.Equal(t, true, SemverExtractStrategy{PinMajor: true, PinMinor: true}.IsCompatible("1.0.0", "1.0.0"))
	assert.Equal(t, true, SemverExtractStrategy{PinMajor: true, PinMinor: true}.IsCompatible("1.0.0", "1.0.1"))
	assert.Equal(t, false, SemverExtractStrategy{PinMajor: true, PinMinor: true}.IsCompatible("1.0.0", "1.1.0"))
	assert.Equal(t, false, SemverExtractStrategy{PinMajor: true, PinMinor: true}.IsCompatible("1.0.0", "2.0.0"))

	assert.Equal(t, true, SemverExtractStrategy{PinMajor: true, PinMinor: true, PinPatch: true}.IsCompatible("1.0.0", "1.0.0"))
	assert.Equal(t, false, SemverExtractStrategy{PinMajor: true, PinMinor: true, PinPatch: true}.IsCompatible("1.0.0", "1.0.1"))
	assert.Equal(t, false, SemverExtractStrategy{PinMajor: true, PinMinor: true, PinPatch: true}.IsCompatible("1.0.0", "1.1.0"))
	assert.Equal(t, false, SemverExtractStrategy{PinMajor: true, PinMinor: true, PinPatch: true}.IsCompatible("1.0.0", "2.0.0"))

	assert.Equal(t, false, SemverExtractStrategy{AllowPrereleases: false}.IsCompatible("1.0.0", "2.0.0-pre"))
	assert.Equal(t, true, SemverExtractStrategy{AllowPrereleases: false}.IsCompatible("1.0.0", "2.0.0"))
	assert.Equal(t, true, SemverExtractStrategy{AllowPrereleases: true}.IsCompatible("1.0.0", "2.0.0-pre"))
}
