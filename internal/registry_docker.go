package internal

import (
	"encoding/json"
	"fmt"
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

var _ Registry = (*DockerRegistry)(nil)

func (r DockerRegistry) GetInterval() time.Duration {
	return r.Interval
}

func (r DockerRegistry) FetchVersions(repository string) ([]string, error) {
	type tagsPageJson struct {
		Name string   `json:"name"`
		Tags []string `json:"tags"`
	}
	client, url := r.createClient()

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

		var tagsPage tagsPageJson
		err = json.Unmarshal(respBody, &tagsPage)
		if err != nil {
			return nil, err
		}
		result = append(result, tagsPage.Tags...)

		nextLink = dockerGetNextLink(resp)
		if strings.HasPrefix(nextLink, "/") {
			nextLink = url + nextLink
		}
	}

	return result, nil
}

func (r DockerRegistry) RetrieveMetadata(repository string, version string) (map[string]string, error) {
	type manifestConfigJson struct {
		MediaType string `json:"mediaType"`
		Digest    string `json:"digest"`
		Size      int64  `json:"size"`
	}
	type manifestLayerJson struct {
		MediaType string `json:"mediaType"`
		Digest    string `json:"digest"`
		Size      int64  `json:"size"`
	}
	type manifestJson struct {
		MediaType     string              `json:"mediaType"`
		SchemaVersion int                 `json:"schemaVersion"`
		Config        manifestConfigJson  `json:"config"`
		Layers        []manifestLayerJson `json:"layers"`
	}
	type manifestListManifestJson struct {
		MediaType string `json:"mediaType"`
		Digest    string `json:"digest"`
		Size      int64  `json:"size"`
		// Platform
	}
	type manifestListJson struct {
		MediaType     string                     `json:"mediaType"`
		SchemaVersion int                        `json:"schemaVersion"`
		Manifests     []manifestListManifestJson `json:"manifests"`
	}
	type configConfigJson struct {
		Labels map[string]string `json:"Labels"`
		// rest omitted
	}
	type configJson struct {
		Config configConfigJson `json:"config"`
		// rest omitted
	}
	client, url := r.createClient()

	req, err := http.NewRequest("GET", url+"/v2/"+repository+"/manifests/"+version, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.list.v2+json, application/vnd.docker.distribution.manifest.v2+json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	manifests := []manifestJson{}

	if resp.Header.Get("content-type") == "application/vnd.docker.distribution.manifest.v2+json" {
		var manifest manifestJson
		err = json.Unmarshal(respBody, &manifest)
		if err != nil {
			return nil, err
		}
		manifests = append(manifests, manifest)
	} else if resp.Header.Get("content-type") == "application/vnd.docker.distribution.manifest.list.v2+json" {
		var manifestList manifestListJson
		err = json.Unmarshal(respBody, &manifestList)
		if err != nil {
			return nil, err
		}
		for _, m := range manifestList.Manifests {
			req2, err := http.NewRequest("GET", url+"/v2/"+repository+"/manifests/"+m.Digest, nil)
			if err != nil {
				return nil, err
			}
			req2.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")
			resp2, err := client.Do(req2)
			if err != nil {
				return nil, err
			}
			defer resp2.Body.Close()
			resp2Body, err := ioutil.ReadAll(resp2.Body)
			if err != nil {
				return nil, err
			}

			var manifest manifestJson
			err = json.Unmarshal(resp2Body, &manifest)
			if err != nil {
				return nil, err
			}
			manifests = append(manifests, manifest)
		}
	} else {
		return nil, fmt.Errorf("unexpected content type %s", req.Header.Get("content-type"))
	}

	result := map[string]string{}

	for _, m := range manifests {
		req3, err := http.NewRequest("GET", url+"/v2/"+repository+"/blobs/"+m.Config.Digest, nil)
		if err != nil {
			return nil, err
		}
		resp3, err := client.Do(req3)
		if err != nil {
			return nil, err
		}
		defer resp3.Body.Close()
		resp3Body, err := ioutil.ReadAll(resp3.Body)
		if err != nil {
			return nil, err
		}
		var config configJson
		err = json.Unmarshal(resp3Body, &config)
		if err != nil {
			return nil, err
		}

		for k, v := range config.Config.Labels {
			result["Label "+k] = v
		}
	}

	return result, nil
}

func (r DockerRegistry) createClient() (http.Client, string) {
	url := strings.TrimSuffix(r.Url, "/")
	username := r.Credentials.Username
	password := r.Credentials.Password
	client := http.Client{
		Transport: dockerWrapTransport(http.DefaultTransport, url, username, password),
	}
	return client, url
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
