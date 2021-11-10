package main

import (
	"testing"
)

func TestLoadGitOpsUpdaterConfig(t *testing.T) {
	yaml := `definitions:
  creds: &creds
    username: user
    password: pass
files:
  includes:
  - '**/*.yaml'
  excludes:
  - '**/*.generated.yaml'
registries:
- name: docker
  interval: 1m
  docker:
    url: https://registry-1.docker.io
    credentials: *creds
- name: helm
  interval: 1h
  helm:
    url: https://charts.helm.sh/stable
    credentials: *creds
policies:
- name: lexicographic
  pattern: '^(?P<all>.*)$'
  extracts:
  - value: '<all>'
    lexicographic: {}
- name: numeric
  pattern: '^(?P<all>.*)$'
  extracts:
  - value: '<all>'
    numeric: {}
- name: semver
  pattern: '^(?P<all>.*)$'
  extracts:
  - value: '<all>'
    semver: {}
`

	c1, r1, p1, err := LoadGitOpsUpdaterConfig([]byte(yaml))
	if err != nil {
		t.Error(err)
		return
	}
	c2 := GitOpsUpdaterConfig{
		Files: GitOpsUpdaterConfigFiles{
			Includes: []string{"**/*.yaml"},
			Excludes: []string{"**/*.generated.yaml"},
		},
		RegistryConfigs: []RegistryConfig{
			{
				Name:     "docker",
				Interval: Duration(60000000000),
				Docker: &RegistryConfigDocker{
					Url: "https://registry-1.docker.io",
					Credentials: HttpBasicCredentials{
						Username: "user",
						Password: "pass",
					},
				},
			},
			{
				Name:     "helm",
				Interval: Duration(3600000000000),
				Helm: &RegistryConfigHelm{
					Url: "https://charts.helm.sh/stable",
					Credentials: HttpBasicCredentials{
						Username: "user",
						Password: "pass",
					},
				},
			},
		},
		PolicyConfigs: []PolicyConfig{
			{
				Name:    "lexicographic",
				Pattern: "^(?P<all>.*)$",
				Extracts: []ExtractConfig{
					{
						Value:         "<all>",
						Lexicographic: &LexicographicExtractStrategyConfig{},
					},
				},
			},
			{
				Name:    "numeric",
				Pattern: "^(?P<all>.*)$",
				Extracts: []ExtractConfig{
					{
						Value:   "<all>",
						Numeric: &NumericExtractStrategyConfig{},
					},
				},
			},
			{
				Name:    "semver",
				Pattern: "^(?P<all>.*)$",
				Extracts: []ExtractConfig{
					{
						Value:  "<all>",
						Semver: &SemverExtractStrategyConfig{},
					},
				},
			},
		},
	}
	if !c1.Equal(c2) {
		t.Errorf("expected %v, got %v", c2, c1)
		return
	}

	ok := false
	_, ok = (*r1)["docker"].(DockerRegistry)
	if !ok {
		t.Errorf("expected DockerRegistry, got %v", (*r1)["docker"])
		return
	}
	_, ok = (*r1)["helm"].(HelmRegistry)
	if !ok {
		t.Errorf("expected HelmRegistry, got %v", (*r1)["helm"])
		return
	}
	_, ok = (*p1)["lexicographic"].Extracts[0].Strategy.(LexicographicExtractStrategy)
	if !ok {
		t.Errorf("expected LexicographicExtractStrategy, got %v", (*p1)["lexicographic"].Extracts[0].Strategy)
		return
	}
	_, ok = (*p1)["numeric"].Extracts[0].Strategy.(NumericExtractStrategy)
	if !ok {
		t.Errorf("expected NumericExtractStrategy, got %v", (*p1)["numeric"].Extracts[0].Strategy)
		return
	}
	_, ok = (*p1)["semver"].Extracts[0].Strategy.(SemverExtractStrategy)
	if !ok {
		t.Errorf("expected SemverExtractStrategy, got %v", (*p1)["semver"].Extracts[0].Strategy)
		return
	}
}
