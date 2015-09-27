package gddoexp

import (
	"sync"
	"time"

	"github.com/golang/gddo/database"
)

// unused stores the time that an unmodified project is considered unused.
const unused = 2 * 365 * 24 * time.Hour

// gddoexpDB contains all used methods from Database type of
// github.com/golang/gddo/database. This is useful for mocking and building
// tests.
type gddoexpDB interface {
	ImporterCount(string) (int, error)
}

// Response stores the information of a path verification on an asynchronous
// check.
type Response struct {
	Path    string
	Archive bool
	Error   error
}

// ShouldArchivePackage determinates if a package should be archived or not.
// It's necessary to inform the GoDoc database to retrieve current stored
// package information.
func ShouldArchivePackage(path string, db gddoexpDB) (bool, error) {
	count, err := db.ImporterCount(path)
	if err != nil {
		return false, NewError(path, ErrorCodeRetrieveImportCounts, err)
	}

	info, err := getGithubInfo(path)
	if err != nil {
		return false, err
	}

	// we only archive the package if there's no reference to it from other
	// projects and if there's no updates in Github on the last 2 years
	archive := count == 0 && time.Now().Sub(info.UpdatedAt) >= unused
	return archive, nil
}

// ShouldArchivePackages determinates if a package should be archived or not,
// but unlike ShouldArchivePackage, it can process a list of packages
// concurrently. It's necessary to inform the GoDoc database to retrieve
// current stored package information.
func ShouldArchivePackages(packages []database.Package, db gddoexpDB) <-chan Response {
	out := make(chan Response)

	var finished struct {
		int
		sync.Mutex
	}

	for _, pkg := range packages {
		go func() {
			archive, err := ShouldArchivePackage(pkg.Path, db)
			out <- Response{
				Path:    pkg.Path,
				Archive: archive,
				Error:   err,
			}

			finished.Lock()
			finished.int++
			if finished.int == len(packages) {
				close(out)
			}
			finished.Unlock()
		}()
	}

	return out
}
