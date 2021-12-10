package internal

import (
	"net/http"
	"strings"
	"time"

	"github.com/heroku/docker-registry-client/registry"
)

type DockerRegistry struct {
	Interval    time.Duration
	Url         string
	Credentials HttpBasicCredentials
}

func (r DockerRegistry) GetInterval() time.Duration {
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
