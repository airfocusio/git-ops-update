package internal

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPolicyParse(t *testing.T) {
	p1 := Policy{}
	actual, err := p1.Parse("v1")
	if assert.NoError(t, err) {
		assert.Equal(t, &[]string{}, actual)
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
			{
				Value:    "<major><minor>",
				Strategy: NumericExtractStrategy{},
			},
		},
	}
	actual, err = p2.Parse("1.2")
	if assert.NoError(t, err) {
		assert.Equal(t, &[]string{"1", "2", "12"}, actual)
	}
}

func TestPolicyFilterAndSort(t *testing.T) {
	p1 := Policy{}
	actual, err := p1.FilterAndSort("1", []string{"1", "2", "3"})
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
	actual, err = p2.FilterAndSort("1.0", strings.Split("18.04 18.10 19.04 19.10 20.04 20.10 21.04 21.10 22.04", " "))
	if assert.NoError(t, err) {
		assert.Equal(t, strings.Split("22.04 21.10 21.04 20.10 20.04 19.10 19.04 18.10 18.04", " "), *actual)
	}
}

func TestPolicyFindNext(t *testing.T) {
	p1 := Policy{}
	actual, err := p1.FindNext("1", []string{"1", "2", "3"})
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
	actual, err = p2.FindNext("1.0", []string{"1.0", "1.1", "2.0", "2.0-pre"})
	if assert.NoError(t, err) {
		assert.Equal(t, "2.0", *actual)
	}
	actual, err = p2.FindNext("1.1", []string{"1.0", "1.1", "2.0", "2.0-pre"})
	if assert.NoError(t, err) {
		assert.Equal(t, "2.0", *actual)
	}
	actual, err = p2.FindNext("2.0", []string{"1.0", "1.1", "2.0", "2.0-pre"})
	if assert.NoError(t, err) {
		assert.Equal(t, "2.0", *actual)
	}
	actual, err = p2.FindNext("3.0", []string{"1.0", "1.1", "2.0", "2.0-pre"})
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
	actual, err = p3.FindNext("1.0.0", []string{"1.0.0", "1.1.0", "2.0.0", "2.0.0-pre"})
	if assert.NoError(t, err) {
		assert.Equal(t, "2.0.0", *actual)
	}
	actual, err = p3.FindNext("1.1.0", []string{"1.0.0", "1.1.0", "2.0.0", "2.0.0-pre"})
	if assert.NoError(t, err) {
		assert.Equal(t, "2.0.0", *actual)
	}
	actual, err = p3.FindNext("2.0.0", []string{"1.0.0", "1.1.0", "2.0.0", "2.0.0-pre"})
	if assert.NoError(t, err) {
		assert.Equal(t, "2.0.0", *actual)
	}
	actual, err = p3.FindNext("3.0.0", []string{"1.0.0", "1.1.0", "2.0.0", "2.0.0-pre"})
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

	assert.Equal(t, true, SemverExtractStrategy{PinMinor: true}.IsCompatible("1.0.0", "1.0.0"))
	assert.Equal(t, true, SemverExtractStrategy{PinMinor: true}.IsCompatible("1.0.0", "1.0.1"))
	assert.Equal(t, false, SemverExtractStrategy{PinMinor: true}.IsCompatible("1.0.0", "1.1.0"))
	assert.Equal(t, false, SemverExtractStrategy{PinMinor: true}.IsCompatible("1.0.0", "2.0.0"))

	assert.Equal(t, true, SemverExtractStrategy{PinPatch: true}.IsCompatible("1.0.0", "1.0.0"))
	assert.Equal(t, false, SemverExtractStrategy{PinPatch: true}.IsCompatible("1.0.0", "1.0.1"))
	assert.Equal(t, false, SemverExtractStrategy{PinPatch: true}.IsCompatible("1.0.0", "1.1.0"))
	assert.Equal(t, false, SemverExtractStrategy{PinPatch: true}.IsCompatible("1.0.0", "2.0.0"))
}
