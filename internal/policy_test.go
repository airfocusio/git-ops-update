package internal

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPolicyParse(t *testing.T) {
	p1 := Policy{}
	actual1, actual2, err := p1.Parse("1", "", "")
	if assert.NoError(t, err) {
		assert.Equal(t, map[string]string{}, actual1)
		assert.Equal(t, []string{}, actual2)
	}

	p2 := Policy{Extracts: []Extract{{Strategy: NumericExtractStrategy{}}}}
	actual1, actual2, err = p2.Parse("1", "", "")
	if assert.NoError(t, err) {
		assert.Equal(t, map[string]string{}, actual1)
		assert.Equal(t, []string{"1"}, actual2)
	}
	actual1, actual2, err = p2.Parse("2", "", "")
	if assert.NoError(t, err) {
		assert.Equal(t, map[string]string{}, actual1)
		assert.Equal(t, []string{"2"}, actual2)
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
	actual1, actual2, err = p3.Parse("1.2", "", "")
	if assert.NoError(t, err) {
		assert.Equal(t, map[string]string{"major": "1", "minor": "2"}, actual1)
		assert.Equal(t, []string{"1", "2", "12"}, actual2)
	}
	actual1, actual2, err = p3.Parse("v1.2-alpine", "v", "-alpine")
	if assert.NoError(t, err) {
		assert.Equal(t, map[string]string{"major": "1", "minor": "2"}, actual1)
		assert.Equal(t, []string{"1", "2", "12"}, actual2)
	}
	actual1, actual2, err = p3.Parse("v1.2", "v", "-alpine")
	if assert.NoError(t, err) {
		assert.Nil(t, actual1)
		assert.Nil(t, actual2)
	}
	actual1, actual2, err = p3.Parse("1.2-alpine", "v", "-alpine")
	if assert.NoError(t, err) {
		assert.Nil(t, actual1)
		assert.Nil(t, actual2)
	}

	p4 := Policy{
		Pattern: regexp.MustCompile(`^(?P<foo_bar>\d+)$`),
		Extracts: []Extract{
			{
				Value:    "<foo_bar>",
				Strategy: NumericExtractStrategy{},
			},
		},
	}
	actual1, actual2, err = p4.Parse("123", "", "")
	if assert.NoError(t, err) {
		assert.Equal(t, map[string]string{"foo_bar": "123"}, actual1)
		assert.Equal(t, []string{"123"}, actual2)
	}

	p5 := Policy{
		Pattern: regexp.MustCompile(`^(?P<major>\d+)(\.(?P<minor>\d+))?$`),
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
	actual1, actual2, err = p5.Parse("1", "", "")
	if assert.NoError(t, err) {
		assert.Equal(t, map[string]string{"major": "1", "minor": ""}, actual1)
		assert.Equal(t, []string{"1", ""}, actual2)
	}
	actual1, actual2, err = p5.Parse("1.2", "", "")
	if assert.NoError(t, err) {
		assert.Equal(t, map[string]string{"major": "1", "minor": "2"}, actual1)
		assert.Equal(t, []string{"1", "2"}, actual2)
	}

	p6 := Policy{
		Pattern: regexp.MustCompile(`^v?(?P<version>.*)$`),
		Extracts: []Extract{
			{
				Key:      "some",
				Value:    "<version>",
				Strategy: SemverExtractStrategy{},
			},
		},
	}
	actual1, actual2, err = p6.Parse("1.2.3-a.b+c.d", "", "")
	if assert.NoError(t, err) {
		assert.Equal(
			t,
			map[string]string{
				"version":      "1.2.3-a.b+c.d",
				"some.major":   "1",
				"some.minor":   "2",
				"some.patch":   "3",
				"some.pre":     "a.b",
				"some.pre.0":   "a",
				"some.pre.1":   "b",
				"some.build":   "c.d",
				"some.build.0": "c",
				"some.build.1": "d",
			},
			actual1,
		)
		assert.Equal(t, []string{"1.2.3-a.b+c.d"}, actual2)
	}
}

func TestPolicyFilterAndSort(t *testing.T) {
	p1 := Policy{}
	actual, err := p1.FilterAndSort("1", []string{"1", "2", "3"}, "", "", nil)
	if assert.NoError(t, err) {
		assert.Equal(t, []string{"3", "2", "1"}, actual)
	}

	p2 := Policy{
		Pattern: regexp.MustCompile(`^(?P<major>\d+)\.(?P<minor>\d+)(-(?P<suffix>.+))?$`),
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
	actual, err = p2.FilterAndSort("1.0", strings.Split("18.04 18.10 19.04 19.10 20.04 20.10 21.04 21.10 22.04", " "), "", "", nil)
	if assert.NoError(t, err) {
		assert.Equal(t, strings.Split("22.04 21.10 21.04 20.10 20.04 19.10 19.04 18.10 18.04", " "), actual)
	}
	actual, err = p2.FilterAndSort("v1.0-ubuntu", strings.Split("17.10 v18.04-ubuntu v18.10-ubuntu v19.04-ubuntu v19.10-ubuntu v20.04-ubuntu v20.10-ubuntu v21.04-ubuntu v21.10-ubuntu v22.04-ubuntu", " "), "v", "-ubuntu", nil)
	if assert.NoError(t, err) {
		assert.Equal(t, strings.Split("v22.04-ubuntu v21.10-ubuntu v21.04-ubuntu v20.10-ubuntu v20.04-ubuntu v19.10-ubuntu v19.04-ubuntu v18.10-ubuntu v18.04-ubuntu", " "), actual)
	}
	_, err = p2.FilterAndSort("v1.0", strings.Split("17.10 v18.04-ubuntu v18.10-ubuntu v19.04-ubuntu v19.10-ubuntu v20.04-ubuntu v20.10-ubuntu v21.04-ubuntu v21.10-ubuntu v22.04-ubuntu", " "), "v", "-ubuntu", nil)
	assert.Error(t, err)
	_, err = p2.FilterAndSort("1.0-ubuntu", strings.Split("17.10 v18.04-ubuntu v18.10-ubuntu v19.04-ubuntu v19.10-ubuntu v20.04-ubuntu v20.10-ubuntu v21.04-ubuntu v21.10-ubuntu v22.04-ubuntu", " "), "v", "-ubuntu", nil)
	assert.Error(t, err)
	actual, err = p2.FilterAndSort("1.0", strings.Split("1.2 1.10 2.1 2.10", " "), "", "", nil)
	if assert.NoError(t, err) {
		assert.Equal(t, strings.Split("2.10 2.1 1.10 1.2", " "), actual)
	}
	actual, err = p2.FilterAndSort("1.0", strings.Split("1.10 1.2 2.10 2.1", " "), "", "", nil)
	if assert.NoError(t, err) {
		assert.Equal(t, strings.Split("2.10 2.1 1.10 1.2", " "), actual)
	}
	actual, err = p2.FilterAndSort("1.0", strings.Split("2.10-a 1.10 1.2 2.10 2.1 1.10-b 1.10-c 1.10-a", " "), "", "", nil)
	if assert.NoError(t, err) {
		assert.Equal(t, strings.Split("2.10-a 2.10 2.1 1.10-c 1.10-b 1.10-a 1.10 1.2", " "), actual)
	}

	p3 := Policy{
		Pattern: regexp.MustCompile(`^(?P<major>\d+)(\.(?P<minor>\d+))?$`),
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
	actual, err = p3.FilterAndSort("1.2", strings.Split("1.0 2 1 1.1 1.2 1.3 2.0", " "), "", "", nil)
	if assert.NoError(t, err) {
		assert.Equal(t, strings.Split("2.0 2 1.3 1.2 1.1 1.0 1", " "), actual)
	}

	p4 := Policy{
		Pattern: regexp.MustCompile(`^((?P<prefix>.*)-)?(?P<major>\d+)\.(?P<minor>\d+)(-(?P<suffix>.+))?$`),
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
	actual, err = p4.FilterAndSort("1.0", strings.Split("2.0 a-2.0-a a-2.0-b b-2.0-a b-2.0-b", " "), "", "", map[string]interface{}{
		"prefix": "a",
	})
	if assert.NoError(t, err) {
		assert.Equal(t, strings.Split("a-2.0-b a-2.0-a", " "), actual)
	}
	actual, err = p4.FilterAndSort("1.0", strings.Split("2.0 a-2.0-a a-2.0-b b-2.0-a b-2.0-b", " "), "", "", map[string]interface{}{
		"prefix": "a",
		"suffix": "a",
	})
	if assert.NoError(t, err) {
		assert.Equal(t, strings.Split("a-2.0-a", " "), actual)
	}
	actual, err = p4.FilterAndSort("1.0", strings.Split("2.0 a-2.0-a a-2.0-b b-2.0-a b-2.0-b", " "), "", "", map[string]interface{}{
		"prefix": []interface{}{"a", ""},
		"suffix": []interface{}{"b", ""},
	})
	if assert.NoError(t, err) {
		assert.Equal(t, strings.Split("a-2.0-b 2.0", " "), actual)
	}
	_, err = p4.FilterAndSort("1.0", strings.Split("2.0 a-2.0-a a-2.0-b b-2.0-a b-2.0-b", " "), "", "", map[string]interface{}{
		"suffix": 23,
	})
	assert.EqualError(t, err, "filter must either be a string or a string list")
	_, err = p4.FilterAndSort("1.0", strings.Split("2.0 a-2.0-a a-2.0-b b-2.0-a b-2.0-b", " "), "", "", map[string]interface{}{
		"suffix": []interface{}{23},
	})
	assert.EqualError(t, err, "filter must either be a string or a string list")

	p5 := Policy{
		Pattern: regexp.MustCompile(`^(?P<version>.*)$`),
		Extracts: []Extract{
			{
				Key:      "key",
				Value:    "<version>",
				Strategy: SemverExtractStrategy{},
			},
		},
	}

	actual, err = p5.FilterAndSort("1.0.0", strings.Split("1.0.0 2.0.0+a 2.0.0+b 2.0.0+c", " "), "", "", map[string]interface{}{
		"key.build": "b",
	})
	if assert.NoError(t, err) {
		assert.Equal(t, strings.Split("2.0.0+b", " "), actual)
	}
	actual, err = p5.FilterAndSort("1.0.0", strings.Split("1.0.0 2.0.0+a 2.0.0+b 2.0.0+c", " "), "", "", map[string]interface{}{
		"key.build": []interface{}{"a", "c"},
	})
	if assert.NoError(t, err) {
		assert.Equal(t, strings.Split("2.0.0+c 2.0.0+a", " "), actual)
	}
}

func TestPolicyFindNext(t *testing.T) {
	p1 := Policy{}
	actual, err := p1.FindNext("1", []string{"1", "2", "3"}, "", "", nil)
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
	actual, err = p2.FindNext("1.0", []string{"1.0", "1.1", "2.0", "2.0-pre"}, "", "", nil)
	if assert.NoError(t, err) {
		assert.Equal(t, "2.0", *actual)
	}
	actual, err = p2.FindNext("1.1", []string{"1.0", "1.1", "2.0", "2.0-pre"}, "", "", nil)
	if assert.NoError(t, err) {
		assert.Equal(t, "2.0", *actual)
	}
	actual, err = p2.FindNext("2.0", []string{"1.0", "1.1", "2.0", "2.0-pre"}, "", "", nil)
	if assert.NoError(t, err) {
		assert.Equal(t, "2.0", *actual)
	}
	actual, err = p2.FindNext("3.0", []string{"1.0", "1.1", "2.0", "2.0-pre"}, "", "", nil)
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
	actual, err = p3.FindNext("1.0.0", []string{"1.0.0", "1.1.0", "2.0.0", "2.0.0-pre"}, "", "", nil)
	if assert.NoError(t, err) {
		assert.Equal(t, "2.0.0", *actual)
	}
	actual, err = p3.FindNext("1.1.0", []string{"1.0.0", "1.1.0", "2.0.0", "2.0.0-pre"}, "", "", nil)
	if assert.NoError(t, err) {
		assert.Equal(t, "2.0.0", *actual)
	}
	actual, err = p3.FindNext("2.0.0", []string{"1.0.0", "1.1.0", "2.0.0", "2.0.0-pre"}, "", "", nil)
	if assert.NoError(t, err) {
		assert.Equal(t, "2.0.0", *actual)
	}
	actual, err = p3.FindNext("3.0.0", []string{"1.0.0", "1.1.0", "2.0.0", "2.0.0-pre"}, "", "", nil)
	if assert.NoError(t, err) {
		assert.Equal(t, "3.0.0", *actual)
	}

	p4 := Policy{
		Pattern: regexp.MustCompile(`^(?P<all>.*)$`),
		Extracts: []Extract{
			{
				Key:      "version",
				Value:    "<all>",
				Strategy: SemverExtractStrategy{},
			},
		},
	}
	actual, err = p4.FindNext("0.10.0", []string{"0.10.0", "0.10.1", "0.10.2", "0.10.3", "0.10.4-pre", "0.11.0-pre"}, "", "", nil)
	if assert.NoError(t, err) {
		assert.Equal(t, "0.10.3", *actual)
	}
	actual, err = p4.FindNext("0.10.4-pre", []string{"0.10.0", "0.10.1", "0.10.2", "0.10.3", "0.10.4-pre", "0.11.0-pre"}, "", "", nil)
	if assert.NoError(t, err) {
		assert.Equal(t, "0.10.3", *actual)
	}
	actual, err = p4.FindNext("0.11.0-pre", []string{"0.10.0", "0.10.1", "0.10.2", "0.10.3", "0.10.4-pre", "0.11.0-pre"}, "", "", nil)
	if assert.NoError(t, err) {
		assert.Equal(t, "0.10.3", *actual)
	}
	actual, err = p4.FindNext("0.12.0-pre", []string{"0.10.0", "0.10.1", "0.10.2", "0.10.3", "0.10.4-pre", "0.11.0-pre"}, "", "", nil)
	if assert.NoError(t, err) {
		assert.Equal(t, "0.10.3", *actual)
	}
	actual, err = p4.FindNext("0.12.0", []string{"0.10.0", "0.10.1", "0.10.2", "0.10.3", "0.10.4-pre", "0.11.0-pre"}, "", "", nil)
	if assert.NoError(t, err) {
		assert.Equal(t, "0.12.0", *actual)
	}
	actual, err = p4.FindNext("0.12.0", []string{"0.10.0", "0.10.1", "0.10.2", "0.10.3", "0.10.4-pre", "0.11.0-pre"}, "", "", map[string]interface{}{
		"version.major": "0",
		"version.minor": "10",
	})
	if assert.NoError(t, err) {
		assert.Equal(t, "0.10.3", *actual)
	}

	p5 := Policy{
		Pattern: regexp.MustCompile(`^(?P<all>.*)$`),
		Extracts: []Extract{
			{
				Key:   "version",
				Value: "<all>",
				Strategy: SemverExtractStrategy{
					AllowPrereleases: true,
				},
			},
		},
	}
	actual, err = p5.FindNext("0.10.0", []string{"0.10.0", "0.10.1", "0.10.2", "0.10.3", "0.10.4-pre", "0.11.0-pre"}, "", "", nil)
	if assert.NoError(t, err) {
		assert.Equal(t, "0.11.0-pre", *actual)
	}
	actual, err = p5.FindNext("0.10.4-pre", []string{"0.10.0", "0.10.1", "0.10.2", "0.10.3", "0.10.4-pre", "0.11.0-pre"}, "", "", nil)
	if assert.NoError(t, err) {
		assert.Equal(t, "0.11.0-pre", *actual)
	}
	actual, err = p5.FindNext("0.11.0-pre", []string{"0.10.0", "0.10.1", "0.10.2", "0.10.3", "0.10.4-pre", "0.11.0-pre"}, "", "", nil)
	if assert.NoError(t, err) {
		assert.Equal(t, "0.11.0-pre", *actual)
	}
	actual, err = p5.FindNext("0.12.0-pre", []string{"0.10.0", "0.10.1", "0.10.2", "0.10.3", "0.10.4-pre", "0.11.0-pre"}, "", "", nil)
	if assert.NoError(t, err) {
		assert.Equal(t, "0.12.0-pre", *actual)
	}
	actual, err = p5.FindNext("0.12.0", []string{"0.10.0", "0.10.1", "0.10.2", "0.10.3", "0.10.4-pre", "0.11.0-pre"}, "", "", nil)
	if assert.NoError(t, err) {
		assert.Equal(t, "0.12.0", *actual)
	}
	actual, err = p5.FindNext("0.12.0", []string{"0.10.0", "0.10.1", "0.10.2", "0.10.3", "0.10.4-pre", "0.11.0-pre"}, "", "", map[string]interface{}{
		"version.major": "0",
		"version.minor": "10",
	})
	if assert.NoError(t, err) {
		assert.Equal(t, "0.10.4-pre", *actual)
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

func TestNumericSortStrategyIsValid(t *testing.T) {
	str := NumericExtractStrategy{}

	assert.Equal(t, true, str.IsValid("0"))
	assert.Equal(t, true, str.IsValid("1"))
	assert.Equal(t, true, str.IsValid("2"))
	assert.Equal(t, false, str.IsValid("-1"))
	assert.Equal(t, false, str.IsValid("a"))
	assert.Equal(t, true, str.IsValid(""))
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

func TestSemverSortStrategyIsValid(t *testing.T) {
	str := SemverExtractStrategy{}
	strRelaxed := SemverExtractStrategy{Relaxed: true}

	assert.Equal(t, true, str.IsValid("0.0.0"))
	assert.Equal(t, true, str.IsValid("1.2.3"))
	assert.Equal(t, true, str.IsValid("1.2.3-rc.1"))
	assert.Equal(t, false, str.IsValid("v1.2.3"))
	assert.Equal(t, false, str.IsValid("1.2"))
	assert.Equal(t, true, strRelaxed.IsValid("1.2"))
	assert.Equal(t, false, str.IsValid("1"))
	assert.Equal(t, true, strRelaxed.IsValid("1"))
	assert.Equal(t, false, str.IsValid(""))
}

func TestSemverSortStrategyCompare(t *testing.T) {
	str := SemverExtractStrategy{}
	strRelaxed := SemverExtractStrategy{Relaxed: true}

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

	assert.Equal(t, 0, strRelaxed.Compare("1", "1.0"))
	assert.Equal(t, 0, strRelaxed.Compare("1.0", "1.0.0"))
	assert.Equal(t, 1, strRelaxed.Compare("1.1", "1.0.9"))
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

func TestSemverSegments(t *testing.T) {
	assert.Equal(t, map[string]string{
		"major": "1",
		"minor": "2",
		"patch": "3",
		"pre":   "",
		"build": "",
	}, SemverExtractStrategy{}.Segments("1.2.3"))

	assert.Equal(t, map[string]string{
		"major":   "1",
		"minor":   "2",
		"patch":   "3",
		"pre":     "a.b",
		"pre.0":   "a",
		"pre.1":   "b",
		"build":   "c.d",
		"build.0": "c",
		"build.1": "d",
	}, SemverExtractStrategy{}.Segments("1.2.3-a.b+c.d"))

	assert.Equal(t, map[string]string{
		"major":   "1",
		"minor":   "2",
		"patch":   "3",
		"pre":     "",
		"build":   "c.d",
		"build.0": "c",
		"build.1": "d",
	}, SemverExtractStrategy{}.Segments("1.2.3+c.d"))
}
