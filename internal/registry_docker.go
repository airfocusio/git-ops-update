package internal

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"regexp"
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
	type tagsPage struct {
		Name string   `json:"name"`
		Tags []string `json:"tags"`
	}

	url := strings.TrimSuffix(r.Url, "/")
	username := r.Credentials.Username
	password := r.Credentials.Password
	client := http.Client{
		Transport: dockerWrapTransport(http.DefaultTransport, url, username, password),
	}

	result := []string{}
	nextLink := url + "/v2/" + repository + "/tags/list"

	for nextLink != "" {
		resp, err := client.Get(nextLink)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		var tagPage tagsPage
		err = json.Unmarshal(respBody, &tagPage)
		if err != nil {
			return nil, err
		}
		result = append(result, tagPage.Tags...)

		nextLink = dockerGetNextLink(resp)
		if strings.HasPrefix(nextLink, "/") {
			nextLink = url + nextLink
		}
	}

	return &result, nil
}

func dockerWrapTransport(transport http.RoundTripper, url, username, password string) http.RoundTripper {
	tokenTransport := &registry.TokenTransport{
		Transport: transport,
		Username:  username,
		Password:  password,
	}
	basicAuthTransport := &registry.BasicTransport{
		Transport: tokenTransport,
		URL:       url,
		Username:  username,
		Password:  password,
	}
	errorTransport := &registry.ErrorTransport{
		Transport: basicAuthTransport,
	}
	return errorTransport
}

func dockerGetNextLink(resp *http.Response) string {
	regex := regexp.MustCompile(`^ *<?([^;>]+)>? *(?:;[^;]*)*; *rel="?next"?(?:;.*)?`)
	for _, link := range resp.Header[http.CanonicalHeaderKey("Link")] {
		parts := regex.FindStringSubmatch(link)
		if parts != nil {
			return parts[1]
		}
	}
	return ""
}
