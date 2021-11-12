package internal

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadGitOpsUpdaterConfig(t *testing.T) {
	yaml := `definitions:
  creds: &creds
    username: user
    password: pass
files:
  includes:
  - '\.yaml$'
  excludes:
  - '\.generated\.yaml$'
registries:
  docker:
    interval: 1m
    docker:
      url: https://registry-1.docker.io
      credentials: *creds
  helm:
    interval: 1h
    helm:
      url: https://charts.helm.sh/stable
      credentials: *creds
policies:
  lexicographic:
    pattern: '^(?P<all>.*)$'
    extracts:
    - value: '<all>'
      lexicographic:
        pin: yes
  numeric:
    pattern: '^(?P<all>.*)$'
    extracts:
    - value: '<all>'
      numeric:
        pin: yes
  semver:
    pattern: '^(?P<all>.*)$'
    extracts:
    - value: '<all>'
      semver:
        pinMajor: yes
        pinMinor: yes
        pinPatch: yes
        allowPrereleases: yes
`

	c1, err := LoadGitOpsUpdaterConfig([]byte(yaml))
	if err != nil {
		t.Error(err)
		return
	}
	c2 := GitOpsUpdaterConfig{
		Files: GitOpsUpdaterConfigFiles{
			Includes: []regexp.Regexp{*regexp.MustCompile(`\.yaml$`)},
			Excludes: []regexp.Regexp{*regexp.MustCompile(`\.generated\.yaml$`)},
		},
		Registries: map[string]Registry{
			"docker": DockerRegistry{
				Interval: Duration(60000000000),
				Url:      "https://registry-1.docker.io",
				Credentials: HttpBasicCredentials{
					Username: "user",
					Password: "pass",
				},
			},
			"helm": HelmRegistry{
				Interval: Duration(3600000000000),
				Url:      "https://charts.helm.sh/stable",
				Credentials: HttpBasicCredentials{
					Username: "user",
					Password: "pass",
				},
			},
		},
		Policies: map[string]Policy{
			"lexicographic": {
				Pattern: regexp.MustCompile(`^(?P<all>.*)$`),
				Extracts: []Extract{
					{
						Value: "<all>",
						Strategy: LexicographicExtractStrategy{
							Pin: true,
						},
					},
				},
			},
			"numeric": {
				Pattern: regexp.MustCompile(`^(?P<all>.*)$`),
				Extracts: []Extract{
					{
						Value: "<all>",
						Strategy: NumericExtractStrategy{
							Pin: true,
						},
					},
				},
			},
			"semver": {
				Pattern: regexp.MustCompile(`^(?P<all>.*)$`),
				Extracts: []Extract{
					{
						Value: "<all>",
						Strategy: SemverExtractStrategy{
							PinMajor:         true,
							PinMinor:         true,
							PinPatch:         true,
							AllowPrereleases: true,
						},
					},
				},
			},
		},
	}
	assert.Equal(t, c1.Files, c2.Files)
	assert.Equal(t, c1.Registries, c2.Registries)
	assert.Equal(t, c1.Policies, c2.Policies)
}
