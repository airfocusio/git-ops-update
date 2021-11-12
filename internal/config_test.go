package internal

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
      lexicographic: {}
  numeric:
    pattern: '^(?P<all>.*)$'
    extracts:
    - value: '<all>'
      numeric: {}
  semver:
    pattern: '^(?P<all>.*)$'
    extracts:
    - value: '<all>'
      semver: {}
`

	c1, err := LoadGitOpsUpdaterConfig([]byte(yaml))
	if err != nil {
		t.Error(err)
		return
	}
	c2 := GitOpsUpdaterConfig{
		Files: GitOpsUpdaterConfigFiles{
			Includes: []string{"**/*.yaml"},
			Excludes: []string{"**/*.generated.yaml"},
		},
		RegistryConfigs: map[string]RegistryConfig{
			"docker": {
				Interval: Duration(60000000000),
				Docker: &RegistryConfigDocker{
					Url: "https://registry-1.docker.io",
					Credentials: HttpBasicCredentials{
						Username: "user",
						Password: "pass",
					},
				},
			},
			"helm": {
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
		PolicyConfigs: map[string]PolicyConfig{
			"lexicographic": {
				Pattern: "^(?P<all>.*)$",
				Extracts: []ExtractConfig{
					{
						Value:         "<all>",
						Lexicographic: &LexicographicExtractStrategyConfig{},
					},
				},
			},
			"numeric": {
				Pattern: "^(?P<all>.*)$",
				Extracts: []ExtractConfig{
					{
						Value:   "<all>",
						Numeric: &NumericExtractStrategyConfig{},
					},
				},
			},
			"semver": {
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
	_, ok = c1.Registries["docker"].(DockerRegistry)
	if !ok {
		t.Errorf("expected DockerRegistry, got %v", c1.Registries["docker"])
		return
	}
	_, ok = c1.Registries["helm"].(HelmRegistry)
	if !ok {
		t.Errorf("expected HelmRegistry, got %v", c1.Registries["helm"])
		return
	}
	_, ok = c1.Policies["lexicographic"].Extracts[0].Strategy.(LexicographicExtractStrategy)
	if !ok {
		t.Errorf("expected LexicographicExtractStrategy, got %v", c1.Policies["lexicographic"].Extracts[0].Strategy)
		return
	}
	_, ok = c1.Policies["numeric"].Extracts[0].Strategy.(NumericExtractStrategy)
	if !ok {
		t.Errorf("expected NumericExtractStrategy, got %v", c1.Policies["numeric"].Extracts[0].Strategy)
		return
	}
	_, ok = c1.Policies["semver"].Extracts[0].Strategy.(SemverExtractStrategy)
	if !ok {
		t.Errorf("expected SemverExtractStrategy, got %v", c1.Policies["semver"].Extracts[0].Strategy)
		return
	}
}
