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
	Files           GitOpsUpdaterConfigFiles `json:"files"`
	RegistryConfigs []RegistryConfig         `json:"registries"`
	PolicyConfigs   []PolicyConfig           `json:"policies"`
}

func LoadGitOpsUpdaterConfig(yaml []byte) (*GitOpsUpdaterConfig, *map[string]Registry, *map[string]Policy, error) {
	config := &GitOpsUpdaterConfig{}

	json, err := utilyaml.ToJSON(yaml)
	if err != nil {
		return nil, nil, nil, err
	}

	if err = utiljson.Unmarshal(json, config); err != nil {
		return nil, nil, nil, err
	}

	registries := map[string]Registry{}
	for _, r := range config.RegistryConfigs {
		if r.Docker != nil {
			registries[r.Name] = DockerRegistry{
				Name:     r.Name,
				Interval: r.Interval,
				Config:   r.Docker,
			}
		} else if r.Helm != nil {
			registries[r.Name] = HelmRegistry{
				Name:     r.Name,
				Interval: r.Interval,
				Config:   r.Helm,
			}
		} else {
			return nil, nil, nil, fmt.Errorf("registry %s is invalid", r.Name)
		}
	}

	policies := map[string]Policy{}
	for _, p := range config.PolicyConfigs {
		extracts := []Extract{}
		for ei, e := range p.Extracts {
			if e.Lexicographic != nil {
				extracts = append(extracts, Extract{Value: e.Value, Strategy: LexicographicExtractStrategy{}})
			} else if e.Numeric != nil {
				extracts = append(extracts, Extract{Value: e.Value, Strategy: NumericExtractStrategy{}})
			} else if e.Semver != nil {
				extracts = append(extracts, Extract{Value: e.Value, Strategy: SemverExtractStrategy{}})
			} else {
				return nil, nil, nil, fmt.Errorf("policy %s strategy %d is invalid", p.Name, ei)
			}
		}
		if len(extracts) == 0 {
			return nil, nil, nil, fmt.Errorf("policy %s has no extracts", p.Name)
		}
		pattern, err := regexp.Compile(p.Pattern)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("policy %s pattern %s is invalid", p.Name, p.Pattern)
		}
		if p.Pattern == "" {
			pattern = nil
		}
		policies[p.Name] = Policy{
			Name:     p.Name,
			Pattern:  pattern,
			Extracts: extracts,
		}
	}

	return config, &registries, &policies, nil
}

func (c1 GitOpsUpdaterConfig) Equal(c2 GitOpsUpdaterConfig) bool {
	return true &&
		cmp.Equal(c1.Files, c2.Files) &&
		cmp.Equal(c1.RegistryConfigs, c2.RegistryConfigs) &&
		cmp.Equal(c1.PolicyConfigs, c2.PolicyConfigs)
}
