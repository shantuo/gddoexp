package gddoexp

import (
	"sync"
	"time"

	"github.com/golang/gddo/database"
	"github.com/juju/ratelimit"
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

// RateLimit controls the number of requests sent to the Github API for
// authenticated and unauthenticated scenarios using the token bucket
// strategy. For more information on the values, please check:
// https://developer.github.com/v3/#rate-limiting
var RateLimit = struct {
	FillInterval     time.Duration
	Capacity         int64
	AuthFillInterval time.Duration
	AuthCapacity     int64
}{
	FillInterval:     time.Minute,
	Capacity:         agents,
	AuthFillInterval: time.Second,
	AuthCapacity:     agents,
}

// gddoDB contains all used methods from Database type of
// github.com/golang/gddo/database. This is useful for mocking and building
// tests.
type gddoDB interface {
	ImporterCount(string) (int, error)
}

// ArchiveResponse stores the information of a path verification on an
// asynchronous check.
type ArchiveResponse struct {
	Path    string
	Archive bool
	Cache   bool
	Error   error
}

// ShouldArchivePackage determinate if a package should be archived or not.
// It's necessary to inform the GoDoc database to retrieve current stored
// package information. An optional argument with the Github authentication
// can be informed to allow more checks per minute in Github API.
func ShouldArchivePackage(p database.Package, db gddoDB, auth *GithubAuth) (archive, cache bool, err error) {
	count, err := db.ImporterCount(p.Path)
	if err != nil {
		// as we didn't perform any request yet, we can return a cache hit to
		// reuse the token
		return false, true, NewError(p.Path, ErrorCodeRetrieveImportCounts, err)
	}

	// don't archive the package if there's a reference to it from other
	// projects (let's avoid send a request to Github if we already no that we
	// don't need to archive it). We return cache hit as no request was made to
	// Github API.
	if count > 0 {
		return false, true, nil
	}

	repository, cache, err := getGithubRepository(p.Path, auth)
	if err != nil {
		return false, cache, err
	}

	// we only archive the package if there's no reference to it from other
	// projects (checked above) and if there's no updates in Github on the last
	// 2 years
	return time.Now().Sub(repository.UpdatedAt) >= unused, cache, nil
}

// ShouldArchivePackages determinate if a package should be archived or not,
// but unlike ShouldArchivePackage, it can process a list of packages
// concurrently. It's necessary to inform the GoDoc database to retrieve
// current stored package information. An optional argument with the Github
// authentication can be informed to allow more checks per minute in Github
// API (we will use token bucket strategy to don't exceed the rate limit).
func ShouldArchivePackages(packages []database.Package, db gddoDB, auth *GithubAuth) <-chan ArchiveResponse {
	out := make(chan ArchiveResponse)

	go func() {
		var bucket *ratelimit.Bucket
		if auth == nil {
			bucket = ratelimit.NewBucket(RateLimit.FillInterval, RateLimit.Capacity)
		} else {
			bucket = ratelimit.NewBucket(RateLimit.AuthFillInterval, RateLimit.AuthCapacity)
		}

		var wg sync.WaitGroup
		wg.Add(agents)

		in := make(chan database.Package)

		for i := 0; i < agents; i++ {
			go func() {
				// if the go routine retrieve a response from cache, it can run again
				// without waiting for a token, as no hit was made in the Github API
				wait := true
				for p := range in {
					if wait {
						bucket.Wait(1)
					} else {
						wait = true
					}

					archive, cache, err := ShouldArchivePackage(p, db, auth)
					out <- ArchiveResponse{
						Path:    p.Path,
						Archive: archive,
						Cache:   cache,
						Error:   err,
					}

					if cache {
						wait = false
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
// small changes for a pull request. An optional argument with the Github
// authentication can be informed to allow more checks per minute in Github API.
func IsFastForkPackage(p database.Package, auth *GithubAuth) (fastFork, cache bool, err error) {
	repository, cacheRepository, err := getGithubRepository(p.Path, auth)
	if err != nil {
		return false, cacheRepository, err
	}

	// if the repository is not a fork we don't need to check the commits
	if !repository.Fork {
		return false, cacheRepository, nil
	}

	commits, cacheCommits, err := getCommits(p.Path, auth)
	if err != nil {
		return false, cacheRepository && cacheCommits, err
	}

	forkLimitDate := repository.CreatedAt.Add(commitsPeriod)
	commitCounts := 0
	fastFork = true

	for _, commit := range commits {
		if commit.Commit.Author.Date.After(forkLimitDate) {
			fastFork = false
			break
		}

		if commit.Commit.Author.Date.After(repository.CreatedAt) {
			commitCounts++
		}
	}

	if commitCounts > commitsLimit {
		fastFork = false
	}

	return fastFork, cacheRepository && cacheCommits, nil
}

// AreFastForkPackages determinate if a package is a fast fork or not,
// but unlike IsFastForkPackage, it can process a list of packages
// concurrently. An optional argument with the Github authentication can be
// informed to allow more checks per minute in Github API (we will use token
// bucket strategy to don't exceed the rate limit).
func AreFastForkPackages(packages []database.Package, auth *GithubAuth) <-chan FastForkResponse {
	out := make(chan FastForkResponse)

	go func() {
		var bucket *ratelimit.Bucket
		if auth == nil {
			bucket = ratelimit.NewBucket(RateLimit.FillInterval, RateLimit.Capacity)
		} else {
			bucket = ratelimit.NewBucket(RateLimit.AuthFillInterval, RateLimit.AuthCapacity)
		}

		var wg sync.WaitGroup
		wg.Add(agents)

		in := make(chan database.Package)

		for i := 0; i < agents; i++ {
			go func() {
				// if the go routine retrieve a response from cache, it can run again
				// without waiting for a token, as no hit was made in the Github API
				wait := true
				for p := range in {
					if wait {
						bucket.Wait(1)
					} else {
						wait = true
					}

					fastFork, cache, err := IsFastForkPackage(p, auth)
					out <- FastForkResponse{
						Path:     p.Path,
						FastFork: fastFork,
						Cache:    cache,
						Error:    err,
					}

					if cache {
						wait = false
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
