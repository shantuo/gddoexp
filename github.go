package gddoexp

import (
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"github.com/gregjones/httpcache"
	"github.com/gregjones/httpcache/diskcache"
)

// Use a global github client for concurrent request.
var GHClient *github.Client

// IsCacheResponse detects if a HTTP response was retrieved from cache or
// not.
var IsCacheResponse func(*http.Response) bool

func init() {
	t := &github.UnauthenticatedRateLimitedTransport{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		Transport: httpcache.NewTransport(
			diskcache.New(path.Join(os.Getenv("HOME"), ".gddoexp")),
		),
	}
	GHClient = github.NewClient(t.Client())
}

func getGithubRepository(path string) (*github.Repository, bool, error) {
	owner, repo := parse(path)
	repository, response, err := GHClient.Repositories.Get(owner, repo)
	if err, ok := err.(*github.RateLimitError); ok {
		t := err.Rate.Reset
		time.Sleep(t.Sub(time.Now()))
		return getGithubRepository(path)
	} else if response != nil && response.Response.StatusCode == 403 {
		time.Sleep(time.Minute)
		return getGithubRepository(path)
	} else if err != nil {
		return nil, true, err
	}

	return repository, IsCacheResponse(response.Response), err
}

// parse split the given GitHub path and return the owner and repo name.
func parse(path string) (string, string) {
	sub := strings.SplitN(path, "/", 4)
	return sub[1], sub[2]
}

// getCommits will retrieve the commits from a Github repository. This function
// also returns if the response was retrieved from a local cache.
func getCommits(path string) ([]github.RepositoryCommit, bool, error) {
	owner, repo := parse(path)
	opt := &github.CommitsListOptions{
		Path:  path,
		Since: time.Now().Add(-unused),
		Until: time.Now(),
	}
	commits, response, err := GHClient.Repositories.ListCommits(owner, repo, opt)
	if err, ok := err.(*github.RateLimitError); ok {
		t := err.Rate.Reset
		time.Sleep(t.Sub(time.Now()))
		return getCommits(path)
	} else if response != nil && response.Response.StatusCode == 403 {
		time.Sleep(time.Minute)
		return getCommits(path)
	} else if err != nil {
		return nil, true, err
	}

	return commits, IsCacheResponse(response.Response), err
}
