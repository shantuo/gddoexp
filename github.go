package gddoexp

import (
	"encoding/json"
	"fmt"
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

// GithubAuth store the authentication information to allow a less
// restrictive rate limit in Github API. Authenticated requests can make up
// to 5000 requests per hour, otherwise you will be limited in 60 requests
// per hour (https://developer.github.com/v3/#rate-limiting).
type GithubAuth struct {
	ID     string
	Secret string
}

// getGithubInfo will retrieve the path project information. For a better
// rate limit the requests must be authenticated, for more information check:
// https://developer.github.com/v3/search/#rate-limit
func getGithubInfo(path string, auth *GithubAuth) (githubInfo, error) {
	var info githubInfo

	if !strings.HasPrefix(path, "github.com/") {
		return info, NewError(path, ErrorCodeNonGithub, nil)
	}

	normalizedPath := strings.TrimPrefix(path, "github.com/")
	normalizedPath = strings.Join(strings.Split(normalizedPath, "/")[:2], "/")

	var url string

	if auth == nil {
		url = "https://api.github.com/repos/" + normalizedPath
	} else {
		url = fmt.Sprintf("https://api.github.com/repos/%s?client_id=%s&client_secret=%s",
			normalizedPath, auth.ID, auth.Secret)
	}

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
