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

// IsCacheResponse detects if a HTTP response was retrieved from cache or
// not.
var IsCacheResponse func(*http.Response) bool

func init() {
	HTTPClient = new(http.Client)
}

// GithubAuth store the authentication information to allow a less
// restrictive rate limit in Github API. Authenticated requests can make up
// to 5000 requests per hour, otherwise you will be limited in 60 requests
// per hour (https://developer.github.com/v3/#rate-limiting).
type GithubAuth struct {
	ID     string
	Secret string
}

// String build the Github authentication in the request query string format.
func (g GithubAuth) String() string {
	return fmt.Sprintf("client_id=%s&client_secret=%s", g.ID, g.Secret)
}

// githubRepository stores the information of a repository. For more
// information check: https://developer.github.com/v3/repos/#get
type githubRepository struct {
	CreatedAt       time.Time `json:"created_at"`
	Fork            bool      `json:"fork"`
	ForksCount      int       `json:"forks_count"`
	NetworkCount    int       `json:"network_count"`
	StargazersCount int       `json:"stargazers_count"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// getGithubRepository will retrieve the path project information. For a
// better rate limit the requests must be authenticated, for more information
// check: https://developer.github.com/v3/search/#rate-limit. This function
// also returns if the response was retrieved from a local cache.
func getGithubRepository(path string, auth *GithubAuth) (repository githubRepository, cache bool, err error) {
	normalizedPath, err := normalizePath(path)
	if err != nil {
		// as we didn't perform any request yet, we can return a cache hit to
		// reuse the token
		return repository, true, err
	}

	url := "https://api.github.com/repos/" + normalizedPath
	if auth != nil {
		url += "?" + auth.String()
	}

	cache, err = doGithub(path, url, &repository)
	return repository, cache, err
}

// githubCommits stores the information of all commits from a repository. For more
// information check:
// https://developer.github.com/v3/repos/commits/#list-commits-on-a-repository.
// Note that there's a difference between the author and the committer:
// http://stackoverflow.com/questions/6755824/what-is-the-difference-between-author-and-committer-in-git
type githubCommits []struct {
	Commit struct {
		Author struct {
			Date time.Time `json:"date"`
		} `json:"author"`
	} `json:"commit"`
}

// getCommits will retrieve the commits from a Github repository. For a
// better rate limit the requests must be authenticated, for more information
// check: https://developer.github.com/v3/search/#rate-limit. This function
// also returns if the response was retrieved from a local cache.
func getCommits(path string, auth *GithubAuth) (commits githubCommits, cache bool, err error) {
	// we don't need to check the error here, because when we retrieved the
	// repository we already checked for it
	normalizedPath, _ := normalizePath(path)

	url := fmt.Sprintf("https://api.github.com/repos/%s/commits", normalizedPath)
	if auth != nil {
		url += "?" + auth.String()
	}

	cache, err = doGithub(path, url, &commits)
	return commits, cache, err
}

// normalizePath identify if the path is from Github and normalize it for the
// Github API request.
func normalizePath(path string) (string, error) {
	if !strings.HasPrefix(path, "github.com/") {
		return "", NewError(path, ErrorCodeNonGithub, nil)
	}

	normalizedPath := strings.TrimPrefix(path, "github.com/")
	normalizedPath = strings.Join(strings.Split(normalizedPath, "/")[:2], "/")
	return normalizedPath, nil
}

// doGithub is the low level function that do actually the work of querying
// Github API and parsing the response. It is also responsible for verifying if
// the response was a cache hit or not.
func doGithub(path, url string, obj interface{}) (cache bool, err error) {
	rsp, err := HTTPClient.Get(url)
	if err != nil {
		return false, NewError(path, ErrorCodeGithubFetch, err)
	}
	defer func() {
		if rsp.Body != nil {
			// for now we aren't checking the error
			rsp.Body.Close()
		}
	}()

	switch rsp.StatusCode {
	case http.StatusOK:
		// valid response
	case http.StatusForbidden:
		return false, NewError(path, ErrorCodeGithubForbidden, nil)
	case http.StatusNotFound:
		return false, NewError(path, ErrorCodeGithubNotFound, nil)
	default:
		return false, NewError(path, ErrorCodeGithubStatusCode, nil)
	}

	decoder := json.NewDecoder(rsp.Body)
	if err := decoder.Decode(&obj); err != nil {
		return false, NewError(path, ErrorCodeGithubParse, err)
	}

	if IsCacheResponse != nil {
		cache = IsCacheResponse(rsp)
	}

	return cache, nil
}
