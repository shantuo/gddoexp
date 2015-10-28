package gddoexp_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/aryann/difflib"
	"github.com/davecgh/go-spew/spew"
	"github.com/golang/gddo/database"
	"github.com/rafaeljusto/gddoexp"
)

func TestShouldArchivePackage(t *testing.T) {
	data := []struct {
		description   string
		path          string
		db            databaseMock
		auth          *gddoexp.GithubAuth
		httpClient    httpClientMock
		expected      bool
		expectedCache bool
		expectedError error
	}{
		{
			description: "it should archive a package (without authentication)",
			path:        "github.com/rafaeljusto/gddoexp",
			db: databaseMock{
				importerCountMock: func(path string) (int, error) {
					return 0, nil
				},
			},
			httpClient: httpClientMock{
				getMock: func(url string) (*http.Response, error) {
					if url != "https://api.github.com/repos/rafaeljusto/gddoexp" {
						return &http.Response{
							StatusCode: http.StatusBadRequest,
						}, nil
					}

					return &http.Response{
						StatusCode: http.StatusOK,
						Body: ioutil.NopCloser(bytes.NewBufferString(`{
  "created_at": "2010-08-03T21:56:23Z",
  "forks_count": 194,
  "network_count": 194,
  "stargazers_count": 1133,
  "updated_at": "` + time.Now().Add(-2*365*24*time.Hour).Format(time.RFC3339) + `"
}`)),
					}, nil
				},
			},
			expected: true,
		},
		{
			description: "it should archive a package from cache (without authentication)",
			path:        "github.com/rafaeljusto/gddoexp",
			db: databaseMock{
				importerCountMock: func(path string) (int, error) {
					return 0, nil
				},
			},
			httpClient: httpClientMock{
				getMock: func(url string) (*http.Response, error) {
					if url != "https://api.github.com/repos/rafaeljusto/gddoexp" {
						return &http.Response{
							StatusCode: http.StatusBadRequest,
						}, nil
					}

					return &http.Response{
						StatusCode: http.StatusOK,
						Header: http.Header{
							"Cache": []string{"1"},
						},
						Body: ioutil.NopCloser(bytes.NewBufferString(`{
  "created_at": "2010-08-03T21:56:23Z",
  "forks_count": 194,
  "network_count": 194,
  "stargazers_count": 1133,
  "updated_at": "` + time.Now().Add(-2*365*24*time.Hour).Format(time.RFC3339) + `"
}`)),
					}, nil
				},
			},
			expected:      true,
			expectedCache: true,
		},
		{
			description: "it should archive a package (authenticated)",
			path:        "github.com/rafaeljusto/gddoexp",
			db: databaseMock{
				importerCountMock: func(path string) (int, error) {
					return 0, nil
				},
			},
			auth: &gddoexp.GithubAuth{
				ID:     "exampleuser",
				Secret: "abc123",
			},
			httpClient: httpClientMock{
				getMock: func(url string) (*http.Response, error) {
					if url != "https://api.github.com/repos/rafaeljusto/gddoexp?client_id=exampleuser&client_secret=abc123" {
						return &http.Response{
							StatusCode: http.StatusBadRequest,
						}, nil
					}

					return &http.Response{
						StatusCode: http.StatusOK,
						Body: ioutil.NopCloser(bytes.NewBufferString(`{
  "created_at": "2010-08-03T21:56:23Z",
  "forks_count": 194,
  "network_count": 194,
  "stargazers_count": 1133,
  "updated_at": "` + time.Now().Add(-2*365*24*time.Hour).Format(time.RFC3339) + `"
}`)),
					}, nil
				},
			},
			expected: true,
		},
		{
			description: "it shouldn't archive a package because of recent commit",
			path:        "github.com/rafaeljusto/gddoexp",
			db: databaseMock{
				importerCountMock: func(path string) (int, error) {
					return 0, nil
				},
			},
			httpClient: httpClientMock{
				getMock: func(url string) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body: ioutil.NopCloser(bytes.NewBufferString(`{
  "created_at": "2010-08-03T21:56:23Z",
  "forks_count": 194,
  "network_count": 194,
  "stargazers_count": 1133,
  "updated_at": "` + time.Now().Add(-2*364*24*time.Hour).Format(time.RFC3339) + `"
}`)),
					}, nil
				},
			},
			expected: false,
		},
		{
			description: "it shouldn't archive a package because of import reference",
			path:        "github.com/rafaeljusto/gddoexp",
			db: databaseMock{
				importerCountMock: func(path string) (int, error) {
					return 1, nil
				},
			},
			httpClient: httpClientMock{
				getMock: func(url string) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body: ioutil.NopCloser(bytes.NewBufferString(`{
  "created_at": "2010-08-03T21:56:23Z",
  "forks_count": 194,
  "network_count": 194,
  "stargazers_count": 1133,
  "updated_at": "` + time.Now().Add(-2*365*24*time.Hour).Format(time.RFC3339) + `"
}`)),
					}, nil
				},
			},
			expected:      false,
			expectedCache: true,
		},
		{
			description: "it should archive a package (project subpath)",
			path:        "github.com/rafaeljusto/gddoexp/cmd/gddoexp",
			db: databaseMock{
				importerCountMock: func(path string) (int, error) {
					return 0, nil
				},
			},
			httpClient: httpClientMock{
				getMock: func(url string) (*http.Response, error) {
					if url != "https://api.github.com/repos/rafaeljusto/gddoexp" {
						return &http.Response{
							StatusCode: http.StatusBadRequest,
						}, nil
					}

					return &http.Response{
						StatusCode: http.StatusOK,
						Body: ioutil.NopCloser(bytes.NewBufferString(`{
  "created_at": "2010-08-03T21:56:23Z",
  "forks_count": 194,
  "network_count": 194,
  "stargazers_count": 1133,
  "updated_at": "` + time.Now().Add(-2*365*24*time.Hour).Format(time.RFC3339) + `"
}`)),
					}, nil
				},
			},
			expected: true,
		},
		{
			description: "it should fail to retrive the import counts from GoDoc database",
			path:        "github.com/rafaeljusto/gddoexp",
			db: databaseMock{
				importerCountMock: func(path string) (int, error) {
					return 0, fmt.Errorf("i'm a crazy error")
				},
			},
			expectedCache: true,
			expectedError: gddoexp.NewError("github.com/rafaeljusto/gddoexp", gddoexp.ErrorCodeRetrieveImportCounts, fmt.Errorf("i'm a crazy error")),
		},
		{
			description: "it should fail when it's not a Github project",
			path:        "bitbucket.org/rafaeljusto/gddoexp",
			db: databaseMock{
				importerCountMock: func(path string) (int, error) {
					return 0, nil
				},
			},
			expectedCache: true,
			expectedError: gddoexp.NewError("bitbucket.org/rafaeljusto/gddoexp", gddoexp.ErrorCodeNonGithub, nil),
		},
		{
			description: "it should fail when there's a HTTP problem with Github API",
			path:        "github.com/rafaeljusto/gddoexp",
			db: databaseMock{
				importerCountMock: func(path string) (int, error) {
					return 0, nil
				},
			},
			httpClient: httpClientMock{
				getMock: func(url string) (*http.Response, error) {
					return nil, fmt.Errorf("i'm a crazy error")
				},
			},
			expectedError: gddoexp.NewError("github.com/rafaeljusto/gddoexp", gddoexp.ErrorCodeGithubFetch, fmt.Errorf("i'm a crazy error")),
		},
		{
			description: "it should fail when the HTTP status code from Github API is 403 Forbidden (ratelimit) without reset HTTP header",
			path:        "github.com/rafaeljusto/gddoexp",
			db: databaseMock{
				importerCountMock: func(path string) (int, error) {
					return 0, nil
				},
			},
			httpClient: httpClientMock{
				getMock: func(url string) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusForbidden,
					}, nil
				},
			},
			expectedError: gddoexp.NewError("github.com/rafaeljusto/gddoexp", gddoexp.ErrorCodeGithubForbidden, nil),
		},
		{
			description: "it should fail when the HTTP status code from Github API is 403 Forbidden (ratelimit) with reset HTTP header",
			path:        "github.com/rafaeljusto/gddoexp",
			db: databaseMock{
				importerCountMock: func(path string) (int, error) {
					return 0, nil
				},
			},
			httpClient: httpClientMock{
				getMock: func() func(string) (*http.Response, error) {
					var requestNumber int

					return func(url string) (*http.Response, error) {
						requestNumber++

						if url != "https://api.github.com/repos/rafaeljusto/gddoexp" {
							return &http.Response{
								StatusCode: http.StatusBadRequest,
							}, nil
						}

						switch requestNumber {
						case 1:
							return &http.Response{
								StatusCode: http.StatusForbidden,
								Header: http.Header{
									"X-Ratelimit-Reset": []string{strconv.FormatInt(time.Now().Add(10*time.Millisecond).Unix(), 10)},
								},
							}, nil

						case 2:
							return &http.Response{
								StatusCode: http.StatusOK,
								Body: ioutil.NopCloser(bytes.NewBufferString(`{
  "created_at": "2010-08-03T21:56:23Z",
  "forks_count": 194,
  "network_count": 194,
  "stargazers_count": 1133,
  "updated_at": "` + time.Now().Add(-2*365*24*time.Hour).Format(time.RFC3339) + `"
}`)),
							}, nil
						}

						return &http.Response{
							StatusCode: http.StatusInternalServerError,
						}, nil
					}
				}(),
			},
			expected: true,
		},
		{
			description: "it should fail when the HTTP status code from Github API is 404 Not Found",
			path:        "github.com/rafaeljusto/gddoexp",
			db: databaseMock{
				importerCountMock: func(path string) (int, error) {
					return 0, nil
				},
			},
			httpClient: httpClientMock{
				getMock: func(url string) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusNotFound,
					}, nil
				},
			},
			expectedError: gddoexp.NewError("github.com/rafaeljusto/gddoexp", gddoexp.ErrorCodeGithubNotFound, nil),
		},
		{
			description: "it should fail when the HTTP status code from Github API isn't valid",
			path:        "github.com/rafaeljusto/gddoexp",
			db: databaseMock{
				importerCountMock: func(path string) (int, error) {
					return 0, nil
				},
			},
			httpClient: httpClientMock{
				getMock: func(url string) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusTeapot,
					}, nil
				},
			},
			expectedError: gddoexp.NewError("github.com/rafaeljusto/gddoexp", gddoexp.ErrorCodeGithubStatusCode, nil),
		},
		{
			description: "it should fail to decode the JSON response from Github",
			path:        "github.com/rafaeljusto/gddoexp",
			db: databaseMock{
				importerCountMock: func(path string) (int, error) {
					return 0, nil
				},
			},
			httpClient: httpClientMock{
				getMock: func(url string) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewBufferString(`{`)),
					}, nil
				},
			},
			expectedError: gddoexp.NewError("github.com/rafaeljusto/gddoexp", gddoexp.ErrorCodeGithubParse, fmt.Errorf("unexpected EOF")),
		},
	}

	httpClientBkp := gddoexp.HTTPClient
	defer func() {
		gddoexp.HTTPClient = httpClientBkp
	}()

	isCacheResponseBkp := gddoexp.IsCacheResponse
	defer func() {
		gddoexp.IsCacheResponse = isCacheResponseBkp
	}()
	gddoexp.IsCacheResponse = func(r *http.Response) bool {
		return r.Header.Get("Cache") == "1"
	}

	for i, item := range data {
		gddoexp.HTTPClient = item.httpClient

		p := database.Package{
			Path: item.path,
		}

		archive, cache, err := gddoexp.ShouldArchivePackage(p, item.db, item.auth)

		if archive != item.expected {
			if item.expected {
				t.Errorf("[%d] %s: expected package to be archived", i, item.description)
			} else {
				t.Errorf("[%d] %s: expected package to don't be archived", i, item.description)
			}
		}

		if cache != item.expectedCache {
			if item.expectedCache {
				t.Errorf("[%d] %s: expected hit in cache", i, item.description)
			} else {
				t.Errorf("[%d] %s: unexpected hit in cache", i, item.description)
			}
		}

		if !reflect.DeepEqual(item.expectedError, err) {
			t.Errorf("[%d] %s: expected error to be “%v” and got “%v”", i, item.description, item.expectedError, err)
		}
	}
}

func TestShouldArchivePackages(t *testing.T) {
	data := []struct {
		description string
		packages    []database.Package
		db          databaseMock
		auth        *gddoexp.GithubAuth
		httpClient  httpClientMock
		expected    []gddoexp.ArchiveResponse
	}{
		{
			description: "it should archive all the packages (without authentication)",
			packages: []database.Package{
				{Path: "github.com/rafaeljusto/gddoexp"},
				{Path: "github.com/golang/gddo"},
				{Path: "github.com/miekg/dns"},
				{Path: "github.com/docker/docker"},
				{Path: "github.com/golang/go"},
			},
			db: databaseMock{
				importerCountMock: func(path string) (int, error) {
					return 0, nil
				},
			},
			httpClient: httpClientMock{
				getMock: func(url string) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body: ioutil.NopCloser(bytes.NewBufferString(`{
  "created_at": "2010-08-03T21:56:23Z",
  "forks_count": 194,
  "network_count": 194,
  "stargazers_count": 1133,
  "updated_at": "` + time.Now().Add(-2*365*24*time.Hour).Format(time.RFC3339) + `"
}`)),
					}, nil
				},
			},
			expected: []gddoexp.ArchiveResponse{
				{
					Path:    "github.com/docker/docker",
					Archive: true,
				},
				{
					Path:    "github.com/golang/gddo",
					Archive: true,
				},
				{
					Path:    "github.com/golang/go",
					Archive: true,
				},
				{
					Path:    "github.com/miekg/dns",
					Archive: true,
				},
				{
					Path:    "github.com/rafaeljusto/gddoexp",
					Archive: true,
				},
			},
		},
		{
			description: "it should archive all the packages (authenticated)",
			packages: []database.Package{
				{Path: "github.com/rafaeljusto/gddoexp"},
				{Path: "github.com/golang/gddo"},
				{Path: "github.com/miekg/dns"},
				{Path: "github.com/docker/docker"},
				{Path: "github.com/golang/go"},
			},
			db: databaseMock{
				importerCountMock: func(path string) (int, error) {
					return 0, nil
				},
			},
			auth: &gddoexp.GithubAuth{
				ID:     "exampleuser",
				Secret: "abc123",
			},
			httpClient: httpClientMock{
				getMock: func(url string) (*http.Response, error) {
					if !strings.HasSuffix(url, "?client_id=exampleuser&client_secret=abc123") {
						return &http.Response{
							StatusCode: http.StatusBadRequest,
						}, nil
					}

					return &http.Response{
						StatusCode: http.StatusOK,
						Body: ioutil.NopCloser(bytes.NewBufferString(`{
  "created_at": "2010-08-03T21:56:23Z",
  "forks_count": 194,
  "network_count": 194,
  "stargazers_count": 1133,
  "updated_at": "` + time.Now().Add(-2*365*24*time.Hour).Format(time.RFC3339) + `"
}`)),
					}, nil
				},
			},
			expected: []gddoexp.ArchiveResponse{
				{
					Path:    "github.com/docker/docker",
					Archive: true,
				},
				{
					Path:    "github.com/golang/gddo",
					Archive: true,
				},
				{
					Path:    "github.com/golang/go",
					Archive: true,
				},
				{
					Path:    "github.com/miekg/dns",
					Archive: true,
				},
				{
					Path:    "github.com/rafaeljusto/gddoexp",
					Archive: true,
				},
			},
		},
		{
			description: "it should archive all the packages (cache hits)",
			packages: []database.Package{
				{Path: "github.com/rafaeljusto/gddoexp"},
				{Path: "github.com/golang/gddo"},
				{Path: "github.com/miekg/dns"},
				{Path: "github.com/docker/docker"},
				{Path: "github.com/golang/go"},
			},
			db: databaseMock{
				importerCountMock: func(path string) (int, error) {
					return 0, nil
				},
			},
			httpClient: httpClientMock{
				getMock: func(url string) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Header: http.Header{
							"Cache": []string{"1"},
						},
						Body: ioutil.NopCloser(bytes.NewBufferString(`{
  "created_at": "2010-08-03T21:56:23Z",
  "forks_count": 194,
  "network_count": 194,
  "stargazers_count": 1133,
  "updated_at": "` + time.Now().Add(-2*365*24*time.Hour).Format(time.RFC3339) + `"
}`)),
					}, nil
				},
			},
			expected: []gddoexp.ArchiveResponse{
				{
					Path:    "github.com/docker/docker",
					Archive: true,
					Cache:   true,
				},
				{
					Path:    "github.com/golang/gddo",
					Archive: true,
					Cache:   true,
				},
				{
					Path:    "github.com/golang/go",
					Archive: true,
					Cache:   true,
				},
				{
					Path:    "github.com/miekg/dns",
					Archive: true,
					Cache:   true,
				},
				{
					Path:    "github.com/rafaeljusto/gddoexp",
					Archive: true,
					Cache:   true,
				},
			},
		},
	}

	httpClientBkp := gddoexp.HTTPClient
	defer func() {
		gddoexp.HTTPClient = httpClientBkp
	}()

	isCacheResponseBkp := gddoexp.IsCacheResponse
	defer func() {
		gddoexp.IsCacheResponse = isCacheResponseBkp
	}()
	gddoexp.IsCacheResponse = func(r *http.Response) bool {
		return r.Header.Get("Cache") == "1"
	}

	for i, item := range data {
		gddoexp.HTTPClient = item.httpClient

		var responses []gddoexp.ArchiveResponse
		for response := range gddoexp.ShouldArchivePackages(item.packages, item.db, item.auth) {
			responses = append(responses, response)
		}

		sort.Sort(byArchiveResponsePath(responses))
		if !reflect.DeepEqual(item.expected, responses) {
			t.Errorf("[%d] %s: mismatch responses.\n%v", i, item.description, diff(item.expected, responses))
		}
	}
}

func TestIsFastForkPackage(t *testing.T) {
	data := []struct {
		description   string
		path          string
		auth          *gddoexp.GithubAuth
		httpClient    httpClientMock
		expected      bool
		expectedCache bool
		expectedError error
	}{
		{
			description: "it should detect a fast fork package (without authentication)",
			path:        "github.com/rafaeljusto/dns",
			httpClient: httpClientMock{
				getMock: func(url string) (*http.Response, error) {
					if url == "https://api.github.com/repos/rafaeljusto/dns" {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body: ioutil.NopCloser(bytes.NewBufferString(`{
  "created_at": "2010-08-03T21:56:23Z",
  "fork": true
}`)),
						}, nil

					} else if url == "https://api.github.com/repos/rafaeljusto/dns/commits" {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body: ioutil.NopCloser(bytes.NewBufferString(`[
  {
    "commit": {
      "author": {
        "date": "2010-08-03T23:00:00Z"
      }
    }
  }
]`)),
						}, nil

					} else {
						return &http.Response{
							StatusCode: http.StatusBadRequest,
						}, nil
					}
				},
			},
			expected: true,
		},
		{
			description: "it should detect a fast fork package (authenticated)",
			path:        "github.com/rafaeljusto/dns",
			auth: &gddoexp.GithubAuth{
				ID:     "exampleuser",
				Secret: "abc123",
			},
			httpClient: httpClientMock{
				getMock: func(url string) (*http.Response, error) {
					if url == "https://api.github.com/repos/rafaeljusto/dns?client_id=exampleuser&client_secret=abc123" {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body: ioutil.NopCloser(bytes.NewBufferString(`{
  "created_at": "2010-08-03T21:56:23Z",
  "fork": true
}`)),
						}, nil

					} else if url == "https://api.github.com/repos/rafaeljusto/dns/commits?client_id=exampleuser&client_secret=abc123" {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body: ioutil.NopCloser(bytes.NewBufferString(`[
  {
    "commit": {
      "author": {
        "date": "2010-08-03T23:00:00Z"
      }
    }
  }
]`)),
						}, nil

					} else {
						return &http.Response{
							StatusCode: http.StatusBadRequest,
						}, nil
					}
				},
			},
			expected: true,
		},
		{
			description: "it should fail when there's a HTTP problem with Github API (repo request)",
			path:        "github.com/rafaeljusto/gddoexp",
			httpClient: httpClientMock{
				getMock: func(url string) (*http.Response, error) {
					return nil, fmt.Errorf("i'm a crazy error")
				},
			},
			expectedError: gddoexp.NewError("github.com/rafaeljusto/gddoexp", gddoexp.ErrorCodeGithubFetch, fmt.Errorf("i'm a crazy error")),
		},
		{
			description: "it should detect that is not a fast fork when the repository isn't a fork",
			path:        "github.com/miekg/dns",
			httpClient: httpClientMock{
				getMock: func(url string) (*http.Response, error) {
					if url == "https://api.github.com/repos/miekg/dns" {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body: ioutil.NopCloser(bytes.NewBufferString(`{
  "created_at": "2010-08-03T21:56:23Z",
  "fork": false
}`)),
						}, nil
					}

					return &http.Response{
						StatusCode: http.StatusBadRequest,
					}, nil
				},
			},
			expected: false,
		},
		{
			description: "it should fail when there's a HTTP problem with Github API (commits request)",
			path:        "github.com/rafaeljusto/dns",
			httpClient: httpClientMock{
				getMock: func(url string) (*http.Response, error) {
					if url == "https://api.github.com/repos/rafaeljusto/dns" {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body: ioutil.NopCloser(bytes.NewBufferString(`{
  "created_at": "2010-08-03T21:56:23Z",
  "fork": true
}`)),
						}, nil
					}

					return nil, fmt.Errorf("i'm a crazy error")
				},
			},
			expectedError: gddoexp.NewError("github.com/rafaeljusto/dns", gddoexp.ErrorCodeGithubFetch, fmt.Errorf("i'm a crazy error")),
		},
		{
			description: "it should detect that is not a fast fork when there's a commit after the commits period",
			path:        "github.com/rafaeljusto/dns",
			httpClient: httpClientMock{
				getMock: func(url string) (*http.Response, error) {
					if url == "https://api.github.com/repos/rafaeljusto/dns" {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body: ioutil.NopCloser(bytes.NewBufferString(`{
  "created_at": "2010-08-03T21:56:23Z",
  "fork": true
}`)),
						}, nil

					} else if url == "https://api.github.com/repos/rafaeljusto/dns/commits" {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body: ioutil.NopCloser(bytes.NewBufferString(`[
  {
    "commit": {
      "author": {
        "date": "2010-08-03T23:00:00Z"
      }
    }
  },
  {
    "commit": {
      "author": {
        "date": "2010-08-11T21:56:24Z"
      }
    }
  }
]`)),
						}, nil

					} else {
						return &http.Response{
							StatusCode: http.StatusBadRequest,
						}, nil
					}
				},
			},
			expected: false,
		},
		{
			description: "it should detect that is not a fast fork when there are too many commits ",
			path:        "github.com/rafaeljusto/dns",
			httpClient: httpClientMock{
				getMock: func(url string) (*http.Response, error) {
					if url == "https://api.github.com/repos/rafaeljusto/dns" {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body: ioutil.NopCloser(bytes.NewBufferString(`{
  "created_at": "2010-08-03T21:56:23Z",
  "fork": true
}`)),
						}, nil

					} else if url == "https://api.github.com/repos/rafaeljusto/dns/commits" {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body: ioutil.NopCloser(bytes.NewBufferString(`[
  {
    "commit": {
      "author": {
        "date": "2010-08-03T23:00:00Z"
      }
    }
  },
  {
    "commit": {
      "author": {
        "date": "2010-08-03T23:01:00Z"
      }
    }
  },
  {
    "commit": {
      "author": {
        "date": "2010-08-03T23:02:00Z"
      }
    }
  }
]`)),
						}, nil

					} else {
						return &http.Response{
							StatusCode: http.StatusBadRequest,
						}, nil
					}
				},
			},
			expected: false,
		},
		{
			description:   "it should fail when it's not a Github project",
			path:          "bitbucket.org/rafaeljusto/gddoexp",
			expectedCache: true,
			expectedError: gddoexp.NewError("bitbucket.org/rafaeljusto/gddoexp", gddoexp.ErrorCodeNonGithub, nil),
		},
	}

	httpClientBkp := gddoexp.HTTPClient
	defer func() {
		gddoexp.HTTPClient = httpClientBkp
	}()

	isCacheResponseBkp := gddoexp.IsCacheResponse
	defer func() {
		gddoexp.IsCacheResponse = isCacheResponseBkp
	}()
	gddoexp.IsCacheResponse = func(r *http.Response) bool {
		return r.Header.Get("Cache") == "1"
	}

	for i, item := range data {
		gddoexp.HTTPClient = item.httpClient

		p := database.Package{
			Path: item.path,
		}

		fastFork, cache, err := gddoexp.IsFastForkPackage(p, item.auth)

		if fastFork != item.expected {
			if item.expected {
				t.Errorf("[%d] %s: expected package to be a fast fork", i, item.description)
			} else {
				t.Errorf("[%d] %s: expected package to don't be a fast fork", i, item.description)
			}
		}

		if cache != item.expectedCache {
			if item.expectedCache {
				t.Errorf("[%d] %s: expected hit in cache", i, item.description)
			} else {
				t.Errorf("[%d] %s: unexpected hit in cache", i, item.description)
			}
		}

		if !reflect.DeepEqual(item.expectedError, err) {
			t.Errorf("[%d] %s: expected error to be “%v” and got “%v”", i, item.description, item.expectedError, err)
		}
	}
}

func TestAreFastForkPackages(t *testing.T) {
	data := []struct {
		description string
		packages    []database.Package
		auth        *gddoexp.GithubAuth
		httpClient  httpClientMock
		expected    []gddoexp.FastForkResponse
	}{
		{
			description: "it should detect that all packages are fast fork (without authentication)",
			packages: []database.Package{
				{Path: "github.com/rafaeljusto/dns"},
				{Path: "github.com/rafaeljusto/go-testdb"},
				{Path: "github.com/rafaeljusto/mysql"},
				{Path: "github.com/rafaeljusto/handy"},
				{Path: "github.com/rafaeljusto/schema"},
			},
			httpClient: httpClientMock{
				getMock: func(url string) (*http.Response, error) {
					if strings.HasSuffix(url, "/commits") {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body: ioutil.NopCloser(bytes.NewBufferString(`[
  {
    "commit": {
      "author": {
        "date": "2010-08-03T23:00:00Z"
      }
    }
  }
]`)),
						}, nil
					}

					return &http.Response{
						StatusCode: http.StatusOK,
						Body: ioutil.NopCloser(bytes.NewBufferString(`{
  "created_at": "2010-08-03T21:56:23Z",
  "fork": true
}`)),
					}, nil
				},
			},
			expected: []gddoexp.FastForkResponse{
				{
					Path:     "github.com/rafaeljusto/dns",
					FastFork: true,
				},
				{
					Path:     "github.com/rafaeljusto/go-testdb",
					FastFork: true,
				},
				{
					Path:     "github.com/rafaeljusto/handy",
					FastFork: true,
				},
				{
					Path:     "github.com/rafaeljusto/mysql",
					FastFork: true,
				},
				{
					Path:     "github.com/rafaeljusto/schema",
					FastFork: true,
				},
			},
		},
		{
			description: "it should detect that all packages are fast fork (authenticated)",
			packages: []database.Package{
				{Path: "github.com/rafaeljusto/dns"},
				{Path: "github.com/rafaeljusto/go-testdb"},
				{Path: "github.com/rafaeljusto/mysql"},
				{Path: "github.com/rafaeljusto/handy"},
				{Path: "github.com/rafaeljusto/schema"},
			},
			auth: &gddoexp.GithubAuth{
				ID:     "exampleuser",
				Secret: "abc123",
			},
			httpClient: httpClientMock{
				getMock: func(url string) (*http.Response, error) {
					if !strings.HasSuffix(url, "?client_id=exampleuser&client_secret=abc123") {
						return &http.Response{
							StatusCode: http.StatusBadRequest,
						}, nil
					}

					url = strings.TrimSuffix(url, "?client_id=exampleuser&client_secret=abc123")

					if strings.HasSuffix(url, "/commits") {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body: ioutil.NopCloser(bytes.NewBufferString(`[
  {
    "commit": {
      "author": {
        "date": "2010-08-03T23:00:00Z"
      }
    }
  }
]`)),
						}, nil
					}

					return &http.Response{
						StatusCode: http.StatusOK,
						Body: ioutil.NopCloser(bytes.NewBufferString(`{
  "created_at": "2010-08-03T21:56:23Z",
  "fork": true
}`)),
					}, nil
				},
			},
			expected: []gddoexp.FastForkResponse{
				{
					Path:     "github.com/rafaeljusto/dns",
					FastFork: true,
				},
				{
					Path:     "github.com/rafaeljusto/go-testdb",
					FastFork: true,
				},
				{
					Path:     "github.com/rafaeljusto/handy",
					FastFork: true,
				},
				{
					Path:     "github.com/rafaeljusto/mysql",
					FastFork: true,
				},
				{
					Path:     "github.com/rafaeljusto/schema",
					FastFork: true,
				},
			},
		},
		{
			description: "it should detect that all packages are fast fork (cache hits)",
			packages: []database.Package{
				{Path: "github.com/rafaeljusto/dns"},
				{Path: "github.com/rafaeljusto/go-testdb"},
				{Path: "github.com/rafaeljusto/mysql"},
				{Path: "github.com/rafaeljusto/handy"},
				{Path: "github.com/rafaeljusto/schema"},
			},
			httpClient: httpClientMock{
				getMock: func(url string) (*http.Response, error) {
					if strings.HasSuffix(url, "/commits") {
						return &http.Response{
							StatusCode: http.StatusOK,
							Header: http.Header{
								"Cache": []string{"1"},
							},
							Body: ioutil.NopCloser(bytes.NewBufferString(`[
  {
    "commit": {
      "author": {
        "date": "2010-08-03T23:00:00Z"
      }
    }
  }
]`)),
						}, nil
					}

					return &http.Response{
						StatusCode: http.StatusOK,
						Header: http.Header{
							"Cache": []string{"1"},
						},
						Body: ioutil.NopCloser(bytes.NewBufferString(`{
  "created_at": "2010-08-03T21:56:23Z",
  "fork": true
}`)),
					}, nil
				},
			},
			expected: []gddoexp.FastForkResponse{
				{
					Path:     "github.com/rafaeljusto/dns",
					FastFork: true,
					Cache:    true,
				},
				{
					Path:     "github.com/rafaeljusto/go-testdb",
					FastFork: true,
					Cache:    true,
				},
				{
					Path:     "github.com/rafaeljusto/handy",
					FastFork: true,
					Cache:    true,
				},
				{
					Path:     "github.com/rafaeljusto/mysql",
					FastFork: true,
					Cache:    true,
				},
				{
					Path:     "github.com/rafaeljusto/schema",
					FastFork: true,
					Cache:    true,
				},
			},
		},
		{
			description: "it should detect that all packages are fast fork (partial cache hits)",
			packages: []database.Package{
				{Path: "github.com/rafaeljusto/dns"},
				{Path: "github.com/rafaeljusto/go-testdb"},
				{Path: "github.com/rafaeljusto/mysql"},
				{Path: "github.com/rafaeljusto/handy"},
				{Path: "github.com/rafaeljusto/schema"},
			},
			httpClient: httpClientMock{
				getMock: func(url string) (*http.Response, error) {
					if strings.HasSuffix(url, "/commits") {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body: ioutil.NopCloser(bytes.NewBufferString(`[
  {
    "commit": {
      "author": {
        "date": "2010-08-03T23:00:00Z"
      }
    }
  }
]`)),
						}, nil
					}

					return &http.Response{
						StatusCode: http.StatusOK,
						Header: http.Header{
							"Cache": []string{"1"},
						},
						Body: ioutil.NopCloser(bytes.NewBufferString(`{
  "created_at": "2010-08-03T21:56:23Z",
  "fork": true
}`)),
					}, nil
				},
			},
			expected: []gddoexp.FastForkResponse{
				{
					Path:     "github.com/rafaeljusto/dns",
					FastFork: true,
				},
				{
					Path:     "github.com/rafaeljusto/go-testdb",
					FastFork: true,
				},
				{
					Path:     "github.com/rafaeljusto/handy",
					FastFork: true,
				},
				{
					Path:     "github.com/rafaeljusto/mysql",
					FastFork: true,
				},
				{
					Path:     "github.com/rafaeljusto/schema",
					FastFork: true,
				},
			},
		},
	}

	httpClientBkp := gddoexp.HTTPClient
	defer func() {
		gddoexp.HTTPClient = httpClientBkp
	}()

	isCacheResponseBkp := gddoexp.IsCacheResponse
	defer func() {
		gddoexp.IsCacheResponse = isCacheResponseBkp
	}()
	gddoexp.IsCacheResponse = func(r *http.Response) bool {
		return r.Header.Get("Cache") == "1"
	}

	for i, item := range data {
		gddoexp.HTTPClient = item.httpClient

		var responses []gddoexp.FastForkResponse
		for response := range gddoexp.AreFastForkPackages(item.packages, item.auth) {
			responses = append(responses, response)
		}

		sort.Sort(byFastForkResponsePath(responses))
		if !reflect.DeepEqual(item.expected, responses) {
			t.Errorf("[%d] %s: mismatch responses.\n%v", i, item.description, diff(item.expected, responses))
		}
	}
}

type databaseMock struct {
	importerCountMock func(string) (int, error)
}

func (d databaseMock) ImporterCount(path string) (int, error) {
	return d.importerCountMock(path)
}

type httpClientMock struct {
	getMock func(string) (*http.Response, error)
}

func (h httpClientMock) Get(url string) (*http.Response, error) {
	return h.getMock(url)
}

func diff(a, b interface{}) []difflib.DiffRecord {
	return difflib.Diff(
		strings.SplitAfter(spew.Sdump(a), "\n"),
		strings.SplitAfter(spew.Sdump(b), "\n"),
	)
}

type byArchiveResponsePath []gddoexp.ArchiveResponse

func (b byArchiveResponsePath) Len() int           { return len(b) }
func (b byArchiveResponsePath) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b byArchiveResponsePath) Less(i, j int) bool { return b[i].Path < b[j].Path }

type byFastForkResponsePath []gddoexp.FastForkResponse

func (b byFastForkResponsePath) Len() int           { return len(b) }
func (b byFastForkResponsePath) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b byFastForkResponsePath) Less(i, j int) bool { return b[i].Path < b[j].Path }
