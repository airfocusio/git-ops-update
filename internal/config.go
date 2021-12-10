package internal

import (
	"fmt"
	"regexp"
	"time"

	"github.com/airfocusio/go-expandenv"
	"gopkg.in/yaml.v3"
)

type RawConfigHttpCredentials struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type RawConfigFiles struct {
	Includes []string `yaml:"includes"`
	Excludes []string `yaml:"excludes"`
}

type RawConfigRegistryDocker struct {
	Interval    time.Duration            `yaml:"interval"`
	Url         string                   `yaml:"url"`
	Credentials RawConfigHttpCredentials `yaml:"credentials"`
}

type RawConfigRegistryHelm struct {
	Interval    time.Duration            `yaml:"interval"`
	Url         string                   `yaml:"url"`
	Credentials RawConfigHttpCredentials `yaml:"credentials"`
}

type RawConfigRegistryGitHubTag struct {
	Interval    time.Duration            `yaml:"interval"`
	Url         string                   `yaml:"url"`
	Credentials RawConfigHttpCredentials `yaml:"credentials"`
}

type RawConfigPolicyExtractLexicographicStrategy struct {
	Value string `yaml:"value"`
	Pin   bool   `yaml:"pin"`
}

type RawConfigPolicyExtractNumericStrategy struct {
	Value string `yaml:"value"`
	Pin   bool   `yaml:"pin"`
}

type RawConfigPolicyExtractSemverStrategy struct {
	Value            string `yaml:"value"`
	PinMajor         bool   `yaml:"pinMajor"`
	PinMinor         bool   `yaml:"pinMinor"`
	PinPatch         bool   `yaml:"pinPatch"`
	AllowPrereleases bool   `yaml:"allowPrereleases"`
}

type RawConfigPolicy struct {
	Pattern  string                   `yaml:"pattern"`
	Extracts []map[string]interface{} `yaml:"extracts"`
}

type RawConfigGitGitHub struct {
	Owner       string `yaml:"owner"`
	Repo        string `yaml:"repo"`
	AccessToken string `yaml:"accessToken"`
}

type RawConfigGitAuthor struct {
	Name  string `yaml:"name"`
	Email string `yaml:"email"`
}

type RawConfigGit struct {
	Author RawConfigGitAuthor  `yaml:"author"`
	GitHub *RawConfigGitGitHub `yaml:"gitHub"`
}

type RawConfig struct {
	Files      RawConfigFiles                    `yaml:"files"`
	Registries map[string]map[string]interface{} `yaml:"registries"`
	Policies   map[string]RawConfigPolicy        `yaml:"policies"`
	Git        RawConfigGit                      `yaml:"git"`
}

type ConfigFiles struct {
	Includes []regexp.Regexp
	Excludes []regexp.Regexp
}

type Config struct {
	Files      ConfigFiles
	Registries map[string]Registry
	Policies   map[string]Policy
	Git        Git
}

func LoadConfig(bytesRaw []byte) (*Config, error) {
	var expansionTemp interface{}
	err := yaml.Unmarshal(bytesRaw, &expansionTemp)
	if err != nil {
		return nil, err
	}
	expansionTemp, err = expandenv.ExpandEnv(expansionTemp)
	if err != nil {
		return nil, err
	}
	bytes, err := yaml.Marshal(expansionTemp)
	if err != nil {
		return nil, err
	}

	config := &RawConfig{}
	err = yaml.Unmarshal(bytes, config)
	if err != nil {
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
		if !validateName(rn) {
			return nil, fmt.Errorf("registry name %s is invalid", rn)
		}
		t, ok := (r["type"]).(string)
		if !ok {
			return nil, fmt.Errorf("registry %s is missing type", rn)
		}
		if t == "docker" {
			rp := RawConfigRegistryDocker{}
			err := decode(r, &rp)
			if err != nil {
				return nil, fmt.Errorf("registry %s is invalid: %w", rn, err)
			}
			registries[rn] = DockerRegistry{
				Interval: rp.Interval,
				Url:      rp.Url,
				Credentials: HttpBasicCredentials{
					Username: rp.Credentials.Username,
					Password: rp.Credentials.Password,
				},
			}
		} else if t == "helm" {
			rp := RawConfigRegistryHelm{}
			err := decode(r, &rp)
			if err != nil {
				return nil, fmt.Errorf("registry %s is invalid: %w", rn, err)
			}
			registries[rn] = HelmRegistry{
				Interval: rp.Interval,
				Url:      rp.Url,
				Credentials: HttpBasicCredentials{
					Username: rp.Credentials.Username,
					Password: rp.Credentials.Password,
				},
			}
		} else if t == "git-hub-tag" {
			rp := RawConfigRegistryGitHubTag{}
			err := decode(r, &rp)
			if err != nil {
				return nil, fmt.Errorf("registry %s is invalid: %w", rn, err)
			}
			registries[rn] = GitHubTagRegistry{
				Interval: rp.Interval,
				Url:      rp.Url,
				Credentials: HttpBasicCredentials{
					Username: rp.Credentials.Username,
					Password: rp.Credentials.Password,
				},
			}
		} else {
			return nil, fmt.Errorf("registry %s has invalid type %s", rn, t)
		}
	}

	policies := map[string]Policy{}
	for pn, p := range config.Policies {
		if !validateName(pn) {
			return nil, fmt.Errorf("policy name %s is invalid", pn)
		}
		extracts := []Extract{}
		for ei, e := range p.Extracts {
			t, ok := (e["type"]).(string)
			if !ok {
				return nil, fmt.Errorf("policy extract %s/%d is missing type", pn, ei)
			}

			if t == "lexicographic" {
				ep := RawConfigPolicyExtractLexicographicStrategy{}
				err := decode(e, &ep)
				if err != nil {
					return nil, fmt.Errorf("policy extract %s/%d is invalid: %w", pn, ei, err)
				}
				extracts = append(extracts, Extract{Value: ep.Value, Strategy: LexicographicExtractStrategy{
					Pin: ep.Pin,
				}})
			} else if t == "numeric" {
				ep := RawConfigPolicyExtractNumericStrategy{}
				err := decode(e, &ep)
				if err != nil {
					return nil, fmt.Errorf("policy extract %s/%d is invalid: %w", pn, ei, err)
				}
				extracts = append(extracts, Extract{Value: ep.Value, Strategy: NumericExtractStrategy{
					Pin: ep.Pin,
				}})
			} else if t == "semver" {
				ep := RawConfigPolicyExtractSemverStrategy{}
				err := decode(e, &ep)
				if err != nil {
					return nil, fmt.Errorf("policy extract %s/%d is invalid: %w", pn, ei, err)
				}
				extracts = append(extracts, Extract{Value: ep.Value, Strategy: SemverExtractStrategy{
					PinMajor:         ep.PinMajor,
					PinMinor:         ep.PinMinor,
					PinPatch:         ep.PinPatch,
					AllowPrereleases: ep.AllowPrereleases,
				}})
			} else {
				return nil, fmt.Errorf("policy %s/%d has invalid type %s", pn, ei, t)
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

	git := Git{}
	gitAuthor := GitAuthor{
		Name:  config.Git.Author.Name,
		Email: config.Git.Author.Email,
	}
	if config.Git.GitHub != nil {
		git.Provider = GitHubGitProvider{
			Author:      gitAuthor,
			AccessToken: config.Git.GitHub.AccessToken,
		}
	} else {
		git.Provider = LocalGitProvider{
			Author: gitAuthor,
		}
	}

	return &Config{
		Files: ConfigFiles{
			Includes: fileIncludes,
			Excludes: fileExcludes,
		},
		Registries: registries,
		Policies:   policies,
		Git:        git,
	}, nil
}

func decode(input interface{}, output interface{}) error {
	bytes, err := yaml.Marshal(input)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(bytes, output)
	if err != nil {
		return err
	}
	return nil
}
