package gddoexp

import "fmt"

// List of possible error codes that can be returned while using this library.
const (
	// ErrorCodeRetrieveImportCounts is used whenever a error occurs while
	// retrieving the import counter from GoDoc database.
	ErrorCodeRetrieveImportCounts ErrorCode = iota

	// ErrorCodeNonGithub is used when the path isn't from Github.
	ErrorCodeNonGithub

	// ErrorCodeGithubFetch is used when there's a problem while retrieving
	// information from Guthub API.
	ErrorCodeGithubFetch

	// ErrorCodeGithubStatusCode is used when the response status code from
	// Github isn't 200 OK.
	ErrorCodeGithubStatusCode

	// ErrorCodeGithubParse is used when there's a problem while parsing the
	// JSON response.
	ErrorCodeGithubParse
)

// ErrorCode stores the type of the error. Useful when we want to perform
// different actions depending on the error type.
type ErrorCode int

// errorCodeMessage translates an error code to an human understandable
// message.
var errorCodeMessage = map[ErrorCode]string{
	ErrorCodeRetrieveImportCounts: "error retrieving import counts",
	ErrorCodeNonGithub:            "not a Github project",
	ErrorCodeGithubFetch:          "error retrieving information from Github",
	ErrorCodeGithubStatusCode:     "unexpected status code from Github",
	ErrorCodeGithubParse:          "error decoding Github response",
}

// Error stores extra information from a low level error indicating the
// context and path the originated the problem.
type Error struct {
	Path    string
	Code    ErrorCode
	Details error
}

// NewError will build a godocexp error.
func NewError(path string, code ErrorCode, details error) Error {
	return Error{
		Path:    path,
		Code:    code,
		Details: details,
	}
}

// Error will show the error in a human readable message.
func (e Error) Error() string {
	if e.Details == nil {
		return fmt.Sprintf("gddoexp: [%s] %s", e.Path, errorCodeMessage[e.Code])
	}

	return fmt.Sprintf("gddoexp: [%s] %s: %s", e.Path, errorCodeMessage[e.Code], e.Details)
}
