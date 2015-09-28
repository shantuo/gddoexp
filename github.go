package gddoexp

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// httpClient contains all used methods to perform requests in Github API.
// This is useful for mocking and building tests.
type httpClient interface {
	Get(string) (*http.Response, error)
}

// HTTPClient is going to be used to perform the Github HTTP requests. We use
// a global variable, as it is safe for concurrent use by multiple goroutines
var HTTPClient httpClient

func init() {
	HTTPClient = new(http.Client)
}

// githubInfo stores the information of a repository. For more information
// check: https://developer.github.com/v3/repos/
type githubInfo struct {
	CreatedAt       time.Time `json:"created_at"`
	ForksCount      int       `json:"forks_count"`
	NetworkCount    int       `json:"network_count"`
	StargazersCount int       `json:"stargazers_count"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// getGithubInfo will retrieve the path project information.
func getGithubInfo(path string) (githubInfo, error) {
	var info githubInfo

	if !strings.HasPrefix(path, "github.com/") {
		return info, NewError(path, ErrorCodeNonGithub, nil)
	}

	normalizedPath := strings.TrimPrefix(path, "github.com/")
	normalizedPath = strings.Join(strings.Split(normalizedPath, "/")[:2], "/")
	url := "https://api.github.com/repos/" + normalizedPath

	rsp, err := HTTPClient.Get(url)
	if err != nil {
		return info, NewError(path, ErrorCodeGithubFetch, err)
	}
	defer func() {
		if rsp.Body != nil {
			// for now we aren't checking the error
			rsp.Body.Close()
		}
	}()

	if rsp.StatusCode != http.StatusOK {
		return info, NewError(path, ErrorCodeGithubStatusCode, nil)
	}

	decoder := json.NewDecoder(rsp.Body)
	if err := decoder.Decode(&info); err != nil {
		return info, NewError(path, ErrorCodeGithubParse, err)
	}

	return info, nil
}
