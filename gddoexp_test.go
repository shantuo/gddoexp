package gddoexp_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/rafaeljusto/gddoexp"
)

func TestShouldArchivePackage(t *testing.T) {
	data := []struct {
		description   string
		path          string
		db            databaseMock
		httpClient    httpClientMock
		expected      bool
		expectedError error
	}{
		{
			description: "it should archive a package",
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
			expected: false,
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
			description: "it should fail when the HTTP status code from Github API isn't OK",
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

	for i, item := range data {
		gddoexp.HTTPClient = item.httpClient

		archive, err := gddoexp.ShouldArchivePackage(item.path, item.db)
		if archive != item.expected {
			if item.expected {
				t.Errorf("[%d] %s: expected package to be archived", i, item.description)
			} else {
				t.Errorf("[%d] %s: expected package to be don't be archived", i, item.description)
			}
		}

		if !reflect.DeepEqual(item.expectedError, err) {
			t.Errorf("[%d] %s: expected error to be “%v” and got “%v”", i, item.description, item.expectedError, err)
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
