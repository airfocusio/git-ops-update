package internal

import (
	"fmt"
	"regexp"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

type RawConfigHttpCredentials struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type RawConfigFiles struct {
	Includes []string `mapstructure:"includes"`
	Excludes []string `mapstructure:"excludes"`
}

type RawConfigRegistryDocker struct {
	Url         string                   `mapstructure:"url"`
	Credentials RawConfigHttpCredentials `mapstructure:"credentials"`
}

type RawConfigRegistryHelm struct {
	Url         string                   `mapstructure:"url"`
	Credentials RawConfigHttpCredentials `mapstructure:"credentials"`
}

type RawConfigRegistryGitHubTag struct {
	Url         string                   `mapstructure:"url"`
	Credentials RawConfigHttpCredentials `mapstructure:"credentials"`
}

type RawConfigRegistry struct {
	Interval  time.Duration               `mapstructure:"interval"`
	Docker    *RawConfigRegistryDocker    `mapstructure:"docker"`
	Helm      *RawConfigRegistryHelm      `mapstructure:"helm"`
	GitHubTag *RawConfigRegistryGitHubTag `mapstructure:"gitHubTag"`
}

type RawConfigPolicyExtractLexicographicStrategy struct {
	Pin bool `mapstructure:"pin"`
}

type RawConfigPolicyExtractNumericStrategyConfig struct {
	Pin bool `mapstructure:"pin"`
}

type RawConfigPolicyExtractSemverStrategy struct {
	PinMajor         bool `mapstructure:"pinMajor"`
	PinMinor         bool `mapstructure:"pinMinor"`
	PinPatch         bool `mapstructure:"pinPatch"`
	AllowPrereleases bool `mapstructure:"allowPrereleases"`
}

type RawConfigPolicyExtract struct {
	Value         string                                       `mapstructure:"value"`
	Lexicographic *RawConfigPolicyExtractLexicographicStrategy `mapstructure:"lexicographic"`
	Numeric       *RawConfigPolicyExtractNumericStrategyConfig `mapstructure:"numeric"`
	Semver        *RawConfigPolicyExtractSemverStrategy        `mapstructure:"semver"`
}

type RawConfigPolicy struct {
	Pattern  string                   `mapstructure:"pattern"`
	Extracts []RawConfigPolicyExtract `mapstructure:"extracts"`
}

type RawConfigGitGitHub struct {
	Owner       string `mapstructure:"owner"`
	Repo        string `mapstructure:"repo"`
	AccessToken string `mapstructure:"accessToken"`
}

type RawConfigGitAuthor struct {
	Name  string `mapstructure:"name"`
	Email string `mapstructure:"email"`
}

type RawConfigGit struct {
	Author RawConfigGitAuthor  `mapstructure:"author"`
	GitHub *RawConfigGitGitHub `mapstructure:"gitHub"`
}

type RawConfig struct {
	Files      RawConfigFiles               `mapstructure:"files"`
	Registries map[string]RawConfigRegistry `mapstructure:"registries"`
	Policies   map[string]RawConfigPolicy   `mapstructure:"policies"`
	Git        RawConfigGit                 `mapstructure:"git"`
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

func LoadConfig(viperInst viper.Viper) (*Config, error) {
	config := &RawConfig{}
	err := viperInst.Unmarshal(&config, viper.DecodeHook(mapstructure.StringToTimeDurationHookFunc()))
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
		} else if r.GitHubTag != nil {
			registries[rn] = GitHubTagRegistry{
				Interval: r.Interval,
				Url:      r.GitHubTag.Url,
				Credentials: HttpBasicCredentials{
					Username: r.GitHubTag.Credentials.Username,
					Password: r.GitHubTag.Credentials.Password,
				},
			}
		} else {
			return nil, fmt.Errorf("registry %s is invalid", rn)
		}
	}

	policies := map[string]Policy{}
	for pn, p := range config.Policies {
		if !validateName(pn) {
			return nil, fmt.Errorf("policy name %s is invalid", pn)
		}
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
