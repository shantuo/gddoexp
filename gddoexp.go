package gddoexp

import (
	"strings"
	"sync"
	"time"

	"github.com/golang/gddo/database"
	"github.com/google/go-github/github"
)

// unused stores the time that an unmodified project is considered unused.
const unused = 2 * 365 * 24 * time.Hour

// commitsLimit is the maximum number of commits made in the fork so we could
// identify as a fast fork.
const commitsLimit = 2

// commitsPeriod is the period after the fork creation date that we will
// consider the commits a fast fork.
const commitsPeriod = 7 * 24 * time.Hour

// agents contains the number of concurrent go routines that will process
// a list of packages
const agents = 4

// gddoDB contains all used methods from Database type of
// github.com/golang/gddo/database. This is useful for mocking and building
// tests.
type gddoDB interface {
	ImporterCount(string) (int, error)
}

// SuppressResponse stores the information of a path verification on an
// asynchronous check.
type SuppressResponse struct {
	Package  database.Package
	Suppress bool
	Cache    bool
	Error    error
}

// ShouldSuppressPackage determinate if a package should be suppressed or not.
// It's necessary to inform the GoDoc database to retrieve current stored
// package information.
func ShouldSuppressPackage(p database.Package, db gddoDB) (suppress, cache bool, err error) {
	if !strings.HasPrefix(p.Path, "github.com") {
		return false, true, NewError(p.Path, ErrorCodeNonGithub, nil)
	}

	count, err := db.ImporterCount(p.Path)
	if err != nil {
		// as we didn't perform any request yet, we can return a cache hit to
		// reuse the token
		return false, true, NewError(p.Path, ErrorCodeRetrieveImportCounts, err)
	}

	// don't suppress the package if there's a reference to it from other
	// projects (let's avoid send a request to Github if we already no that we
	// don't need to suppress it). We return cache hit as no request was made to
	// Github API.
	if count > 0 {
		return false, true, nil
	}

	repository, cacheRepository, err := getGithubRepository(p.Path)
	if err != nil {
		return false, cacheRepository, err
	}

	// we only suppress the package if there's no reference to it from other
	// projects (checked above) and if there's no updates in Github on the last
	// 2 years
	if time.Now().Sub(repository.UpdatedAt.Time) >= unused {
		return true, cacheRepository, nil
	}

	// we will check if the package is a fork with a few commits for a pull
	// request, if so we consider it a fast fork and is eligible to be
	// suppressed
	fastFork, cacheFastFork, err := isFastForkPackage(p, repository)
	return fastFork, cacheRepository && cacheFastFork, err
}

// ShouldSuppressPackages determinate if a package should be suppressed or not,
// but unlike ShouldSuppressPackage, it can process a list of packages
// concurrently. It's necessary to inform the GoDoc database to retrieve
// current stored package information.
func ShouldSuppressPackages(packages []database.Package, db gddoDB) <-chan SuppressResponse {
	out := make(chan SuppressResponse, agents)

	go func() {
		var wg sync.WaitGroup
		wg.Add(agents)

		in := make(chan database.Package)

		for i := 0; i < agents; i++ {
			go func() {
				for p := range in {
					suppress, cache, err := ShouldSuppressPackage(p, db)
					out <- SuppressResponse{
						Package:  p,
						Suppress: suppress,
						Cache:    cache,
						Error:    err,
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

// FastForkResponse stores the information of a path verification on an
// asynchronous check.
type FastForkResponse struct {
	Path     string
	FastFork bool
	Cache    bool
	Error    error
}

// IsFastForkPackage identifies if a package is a fork created only to make
// small changes for a pull request.
func IsFastForkPackage(p database.Package) (fastFork, cache bool, err error) {
	repository, cacheRepository, err := getGithubRepository(p.Path)
	if err != nil {
		return false, cacheRepository, err
	}

	fastFork, cache, err = isFastForkPackage(p, repository)
	return fastFork, cache && cacheRepository, err
}

// isFastForkPackage is the low level function that will actually check if
// the package is a fast fork. It receives the repository information so we can
// reuse it with the function ShouldSuppressPackage.
func isFastForkPackage(p database.Package, repository *github.Repository) (fastFork, cache bool, err error) {
	// if the repository is not a fork we don't need to check the commits
	if !*repository.Fork {
		return false, true, nil
	}

	commits, cache, err := getCommits(p.Path)
	if err != nil {
		return false, cache, err
	}

	forkLimitDate := repository.CreatedAt.Add(commitsPeriod)
	commitCounts := 0
	fastFork = true

	for _, commit := range commits {
		if commit.Commit.Author.Date.After(forkLimitDate) {
			fastFork = false
			break
		}

		if commit.Commit.Author.Date.After(repository.CreatedAt.Time) {
			commitCounts++
		}
	}

	if commitCounts > commitsLimit {
		fastFork = false
	}

	return fastFork, cache, nil
}

// AreFastForkPackages determinate if a package is a fast fork or not,
// but unlike IsFastForkPackage, it can process a list of packages
// concurrently.
func AreFastForkPackages(packages []database.Package) <-chan FastForkResponse {
	out := make(chan FastForkResponse, agents)

	go func() {
		var wg sync.WaitGroup
		wg.Add(agents)

		in := make(chan database.Package)

		for i := 0; i < agents; i++ {
			go func() {
				for p := range in {
					fastFork, cache, err := IsFastForkPackage(p)
					out <- FastForkResponse{
						Path:     p.Path,
						FastFork: fastFork,
						Cache:    cache,
						Error:    err,
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
