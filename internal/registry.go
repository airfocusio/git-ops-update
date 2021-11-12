package internal

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/heroku/docker-registry-client/registry"
)

type HttpBasicCredentials struct {
	Username string
	Password string
}

type Registry interface {
	GetInterval() Duration
	FetchVersions(resource string) (*[]string, error)
}

type DockerRegistry struct {
	Interval    Duration
	Url         string
	Credentials HttpBasicCredentials
}

type HelmRegistry struct {
	Interval    Duration
	Url         string
	Credentials HttpBasicCredentials
}

func (r DockerRegistry) GetInterval() Duration {
	return r.Interval
}

func (r DockerRegistry) FetchVersions(repository string) (*[]string, error) {
	url := strings.TrimSuffix(r.Url, "/")
	username := r.Credentials.Username
	password := r.Credentials.Password
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

func (r HelmRegistry) GetInterval() Duration {
	return r.Interval
}

func (r HelmRegistry) FetchVersions(chart string) (*[]string, error) {
	url := strings.TrimSuffix(r.Url, "/") + "/index.yaml"
	username := r.Credentials.Username
	password := r.Credentials.Password
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
	err = readYaml(body, &index)
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
