package internal

import (
	utiljson "encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"

	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
)

type GitOpsUpdaterConfigRawHttpCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type GitOpsUpdaterConfigRawFiles struct {
	Includes []string `json:"includes"`
	Excludes []string `json:"excludes"`
}

type GitOpsUpdaterConfigRawRegistryDocker struct {
	Url         string                                `json:"url"`
	Credentials GitOpsUpdaterConfigRawHttpCredentials `json:"credentials"`
}

type GitOpsUpdaterConfigRawRegistryHelm struct {
	Url         string                                `json:"url"`
	Credentials GitOpsUpdaterConfigRawHttpCredentials `json:"credentials"`
}

type GitOpsUpdaterConfigRawRegistry struct {
	Interval Duration                              `json:"interval"`
	Docker   *GitOpsUpdaterConfigRawRegistryDocker `json:"docker"`
	Helm     *GitOpsUpdaterConfigRawRegistryHelm   `json:"helm"`
}

type GitOpsUpdaterConfigRawPolicyExtractLexicographicStrategy struct {
	Pin bool `json:"pin"`
}

type GitOpsUpdaterConfigRawPolicyExtractNumericStrategyConfig struct {
	Pin bool `json:"pin"`
}

type GitOpsUpdaterConfigRawPolicyExtractSemverStrategy struct {
	PinMajor         bool `json:"pinMajor"`
	PinMinor         bool `json:"pinMinor"`
	PinPatch         bool `json:"pinPatch"`
	AllowPrereleases bool `json:"allowPrereleases"`
}

type GitOpsUpdaterConfigRawPolicyExtract struct {
	Value         string                                                    `json:"value"`
	Lexicographic *GitOpsUpdaterConfigRawPolicyExtractLexicographicStrategy `json:"lexicographic"`
	Numeric       *GitOpsUpdaterConfigRawPolicyExtractNumericStrategyConfig `json:"numeric"`
	Semver        *GitOpsUpdaterConfigRawPolicyExtractSemverStrategy        `json:"semver"`
}

type GitOpsUpdaterConfigRawPolicy struct {
	Pattern  string                                `json:"pattern"`
	Extracts []GitOpsUpdaterConfigRawPolicyExtract `json:"extracts"`
}

type GitOpsUpdaterConfigRaw struct {
	Files      GitOpsUpdaterConfigRawFiles               `json:"files"`
	Registries map[string]GitOpsUpdaterConfigRawRegistry `json:"registries"`
	Policies   map[string]GitOpsUpdaterConfigRawPolicy   `json:"policies"`
}

type GitOpsUpdaterConfigFiles struct {
	Includes []regexp.Regexp
	Excludes []regexp.Regexp
}

type GitOpsUpdaterConfig struct {
	Files      GitOpsUpdaterConfigFiles
	Registries map[string]Registry
	Policies   map[string]Policy
}

func LoadGitOpsUpdaterConfig(yaml []byte) (*GitOpsUpdaterConfig, error) {
	config := &GitOpsUpdaterConfigRaw{}

	json, err := utilyaml.ToJSON(yaml)
	if err != nil {
		return nil, err
	}

	if err = utiljson.Unmarshal(json, config); err != nil {
		return nil, err
	}

	fileIncludes := []regexp.Regexp{}
	for _, i := range config.Files.Includes {
		regex, err := regexp.Compile(i)
		if err != nil {
			return nil, err
		}
		fileIncludes = append(fileIncludes, *regex)
	}

	fileExcludes := []regexp.Regexp{}
	for _, e := range config.Files.Excludes {
		regex, err := regexp.Compile(e)
		if err != nil {
			return nil, err
		}
		fileExcludes = append(fileExcludes, *regex)
	}

	registries := map[string]Registry{}
	for rn, r := range config.Registries {
		if r.Docker != nil {
			registries[rn] = DockerRegistry{
				Interval: r.Interval,
				Url:      r.Docker.Url,
				Credentials: HttpBasicCredentials{
					Username: r.Docker.Credentials.Username,
					Password: r.Docker.Credentials.Password,
				},
			}
		} else if r.Helm != nil {
			registries[rn] = HelmRegistry{
				Interval: r.Interval,
				Url:      r.Helm.Url,
				Credentials: HttpBasicCredentials{
					Username: r.Helm.Credentials.Username,
					Password: r.Helm.Credentials.Password,
				},
			}
		} else {
			return nil, fmt.Errorf("registry %s is invalid", rn)
		}
	}

	policies := map[string]Policy{}
	for pn, p := range config.Policies {
		extracts := []Extract{}
		for ei, e := range p.Extracts {
			if e.Lexicographic != nil {
				extracts = append(extracts, Extract{Value: e.Value, Strategy: LexicographicExtractStrategy{
					Pin: e.Lexicographic.Pin,
				}})
			} else if e.Numeric != nil {
				extracts = append(extracts, Extract{Value: e.Value, Strategy: NumericExtractStrategy{
					Pin: e.Numeric.Pin,
				}})
			} else if e.Semver != nil {
				extracts = append(extracts, Extract{Value: e.Value, Strategy: SemverExtractStrategy{
					PinMajor:         e.Semver.PinMajor,
					PinMinor:         e.Semver.PinMinor,
					PinPatch:         e.Semver.PinPatch,
					AllowPrereleases: e.Semver.AllowPrereleases,
				}})
			} else {
				return nil, fmt.Errorf("policy %s strategy %d is invalid", pn, ei)
			}
		}
		if len(extracts) == 0 {
			return nil, fmt.Errorf("policy %s has no extracts", pn)
		}
		pattern, err := regexp.Compile(p.Pattern)
		if err != nil {
			return nil, fmt.Errorf("policy %s pattern %s is invalid", pn, p.Pattern)
		}
		if p.Pattern == "" {
			pattern = nil
		}
		policies[pn] = Policy{
			Pattern:  pattern,
			Extracts: extracts,
		}
	}

	return &GitOpsUpdaterConfig{
		Files: GitOpsUpdaterConfigFiles{
			Includes: fileIncludes,
			Excludes: fileExcludes,
		},
		Registries: registries,
		Policies:   policies,
	}, nil
}

func LoadGitOpsUpdaterConfigFromFile(file string) (*GitOpsUpdaterConfig, error) {
	configRaw, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	config, err := LoadGitOpsUpdaterConfig(configRaw)
	if err != nil {
		return nil, err
	}
	return config, nil
}
