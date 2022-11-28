package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var _ Registry = (*GitHubTagRegistry)(nil)

type GitHubTagRegistry struct {
	Interval    time.Duration
	Url         string
	Credentials HttpBasicCredentials
}

type gitHubTagRegistryRef struct {
	Ref    string `json:"ref"`
	NodeId string `json:"node_id"`
	Url    string `json:"url"`
}

func (r GitHubTagRegistry) GetInterval() time.Duration {
	return r.Interval
}

func (r GitHubTagRegistry) FetchVersions(repository string) ([]string, error) {
	LogDebug("Fetching versions from github-tag registry %s", repository)
	baseUrl := "https://api.github.com"
	if r.Url != "" {
		baseUrl = strings.TrimSuffix(r.Url, "/")
	}
	url := fmt.Sprintf("%s/repos/%s/git/matching-refs/tags", baseUrl, repository)

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
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	refs := []gitHubTagRegistryRef{}
	err = json.Unmarshal(body, &refs)
	if err != nil {
		return nil, err
	}

	result := []string{}
	for _, ref := range refs {
		if strings.HasPrefix(ref.Ref, "refs/tags/") {
			result = append(result, strings.TrimPrefix(ref.Ref, "refs/tags/"))
		}
	}

	return result, nil
}
