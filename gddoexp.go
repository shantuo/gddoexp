package gddoexp

import (
	"sync"
	"time"

	"github.com/golang/gddo/database"
)

// unused stores the time that an unmodified project is considered unused.
const unused = 2 * 365 * 24 * time.Hour

// agents contains the number of concurrent go routines that will process
// a list of packages
const agents = 10

// gddoDB contains all used methods from Database type of
// github.com/golang/gddo/database. This is useful for mocking and building
// tests.
type gddoDB interface {
	ImporterCount(string) (int, error)
}

// Response stores the information of a path verification on an asynchronous
// check.
type Response struct {
	Path    string
	Archive bool
	Error   error
}

// ShouldArchivePackage determinate if a package should be archived or not.
// It's necessary to inform the GoDoc database to retrieve current stored
// package information. An optional argument with the Github authentication
// can be informed to allow more checks per minute in Github API.
func ShouldArchivePackage(p database.Package, db gddoDB, auth *GithubAuth) (bool, error) {
	count, err := db.ImporterCount(p.Path)
	if err != nil {
		return false, NewError(p.Path, ErrorCodeRetrieveImportCounts, err)
	}

	info, err := getGithubInfo(p.Path, auth)
	if err != nil {
		return false, err
	}

	// we only archive the package if there's no reference to it from other
	// projects and if there's no updates in Github on the last 2 years
	archive := count == 0 && time.Now().Sub(info.UpdatedAt) >= unused
	return archive, nil
}

// ShouldArchivePackages determinate if a package should be archived or not,
// but unlike ShouldArchivePackage, it can process a list of packages
// concurrently. It's necessary to inform the GoDoc database to retrieve
// current stored package information. An optional argument with the Github
// authentication can be informed to allow more checks per minute in Github
// API.
func ShouldArchivePackages(packages []database.Package, db gddoDB, auth *GithubAuth) <-chan Response {
	out := make(chan Response)

	go func() {
		var wg sync.WaitGroup
		wg.Add(agents)

		in := make(chan database.Package)

		for i := 0; i < agents; i++ {
			go func() {
				for p := range in {
					archive, err := ShouldArchivePackage(p, db, auth)
					out <- Response{
						Path:    p.Path,
						Archive: archive,
						Error:   err,
					}
				}

				wg.Done()
			}()
		}

		for _, pkg := range packages {
			in <- pkg
		}

		close(in)
		wg.Wait()
		close(out)
	}()

	return out
}
