package gddoexp

import "time"

// unused stores the time that an unmodified project is considered unused.
const unused = 2 * 365 * 24 * time.Hour

// database contains all used methods from Database type of
// github.com/golang/gddo/database. This is useful for mocking and building
// tests.
type database interface {
	ImporterCount(string) (int, error)
}

// ShouldArchivePackage determinates if a package should be archived or not.
// It's necessary to inform the GoDoc database to retrieve current stored
// package information.
func ShouldArchivePackage(path string, db database) (bool, error) {
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
