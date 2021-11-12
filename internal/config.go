package internal

import (
	utiljson "encoding/json"
	"fmt"
	"regexp"

	"github.com/google/go-cmp/cmp"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
)

type GitOpsUpdaterConfigFiles struct {
	Includes []string `json:"includes"`
	Excludes []string `json:"excludes"`
}

type GitOpsUpdaterConfig struct {
	Files           GitOpsUpdaterConfigFiles  `json:"files"`
	RegistryConfigs map[string]RegistryConfig `json:"registries"`
	PolicyConfigs   map[string]PolicyConfig   `json:"policies"`
	Registries      map[string]Registry
	Policies        map[string]Policy
}

func LoadGitOpsUpdaterConfig(yaml []byte) (*GitOpsUpdaterConfig, error) {
	config := &GitOpsUpdaterConfig{}

	json, err := utilyaml.ToJSON(yaml)
	if err != nil {
		return nil, err
	}

	if err = utiljson.Unmarshal(json, config); err != nil {
		return nil, err
	}

	registries := map[string]Registry{}
	for rn, r := range config.RegistryConfigs {
		if r.Docker != nil {
			registries[rn] = DockerRegistry{
				Interval: r.Interval,
				Config:   r.Docker,
			}
		} else if r.Helm != nil {
			registries[rn] = HelmRegistry{
				Interval: r.Interval,
				Config:   r.Helm,
			}
		} else {
			return nil, fmt.Errorf("registry %s is invalid", rn)
		}
	}

	policies := map[string]Policy{}
	for pn, p := range config.PolicyConfigs {
		extracts := []Extract{}
		for ei, e := range p.Extracts {
			if e.Lexicographic != nil {
				extracts = append(extracts, Extract{Value: e.Value, Strategy: LexicographicExtractStrategy{}})
			} else if e.Numeric != nil {
				extracts = append(extracts, Extract{Value: e.Value, Strategy: NumericExtractStrategy{}})
			} else if e.Semver != nil {
				extracts = append(extracts, Extract{Value: e.Value, Strategy: SemverExtractStrategy{}})
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

	config.Registries = registries
	config.Policies = policies

	return config, nil
}

func (c1 GitOpsUpdaterConfig) Equal(c2 GitOpsUpdaterConfig) bool {
	return true &&
		cmp.Equal(c1.Files, c2.Files) &&
		cmp.Equal(c1.RegistryConfigs, c2.RegistryConfigs) &&
		cmp.Equal(c1.PolicyConfigs, c2.PolicyConfigs)
}
