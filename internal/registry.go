package internal

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/heroku/docker-registry-client/registry"
	"gopkg.in/yaml.v3"
)

type HttpBasicCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Registry interface {
	FetchVersions(resource string) (*[]string, error)
}

type RegistryConfigDocker struct {
	Url         string               `json:"url"`
	Credentials HttpBasicCredentials `json:"credentials"`
}

type RegistryConfigHelm struct {
	Url         string               `json:"url"`
	Credentials HttpBasicCredentials `json:"credentials"`
}

type RegistryConfig struct {
	Name     string                `json:"name"`
	Interval Duration              `json:"interval"`
	Docker   *RegistryConfigDocker `json:"docker"`
	Helm     *RegistryConfigHelm   `json:"helm"`
}

type DockerRegistry struct {
	Name     string
	Interval Duration
	Config   *RegistryConfigDocker
}

type HelmRegistry struct {
	Name     string
	Interval Duration
	Config   *RegistryConfigHelm
}

func (r DockerRegistry) FetchVersions(repository string) (*[]string, error) {
	url := strings.TrimSuffix(r.Config.Url, "/")
	username := r.Config.Credentials.Username
	password := r.Config.Credentials.Password
	transport := registry.WrapTransport(http.DefaultTransport, url, username, password)
	client := &registry.Registry{
		URL: url,
		Client: &http.Client{
			Transport: transport,
		},
		Logf: registry.Quiet,
	}
	tags, err := client.Tags(repository)
	if err != nil {
		return nil, err
	}
	return &tags, nil
}

type helmRegistryIndex struct {
	ApiVersion string `yaml:"apiVersion"`
	Entries    map[string][]struct {
		ApiVersion string `yaml:"apiVersion"`
		AppVersion string `yaml:"appVersion"`
		Name       string `yaml:"name"`
		Version    string `yaml:"version"`
	} `yaml:"entries"`
}

func (r HelmRegistry) FetchVersions(chart string) (*[]string, error) {
	url := strings.TrimSuffix(r.Config.Url, "/") + "/index.yaml"
	username := r.Config.Credentials.Username
	password := r.Config.Credentials.Password
	req, err := http.NewRequest("GET", url, nil)
	client := &http.Client{}
	if err != nil {
		return nil, err
	}
	if username != "" {
		req.SetBasicAuth(username, password)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	index := helmRegistryIndex{}
	err = yaml.Unmarshal(body, &index)
	if err != nil {
		return nil, err
	}

	versions, ok := index.Entries[chart]
	if !ok {
		return nil, fmt.Errorf("chart %s could not be found", chart)
	}
	result := []string{}
	for _, version := range versions {
		result = append(result, version.Version)
	}

	return &result, nil
}
