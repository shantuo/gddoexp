package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/cheggaaa/pb"
	"github.com/golang/gddo/database"
	"github.com/gregjones/httpcache"
	"github.com/gregjones/httpcache/diskcache"
	"github.com/rafaeljusto/gddoexp"
)

func init() {
	gddoexp.IsCacheResponse = func(r *http.Response) bool {
		return r.Header.Get(httpcache.XFromCache) == "1"
	}
}

func main() {
	clientID := flag.String("id", "", "Github client ID")
	clientSecret := flag.String("secret", "", "Github client secret")
	output := flag.String("output", "gddoexp.out", "Output file")
	progress := flag.Bool("progress", false, "Show a progress bar")
	flag.Parse()

	var auth *gddoexp.GithubAuth
	if (clientID != nil && *clientID != "") || (clientSecret != nil && *clientSecret != "") {
		if *clientID == "" || *clientSecret == "" {
			fmt.Println("to enable Gthub authentication, you need to inform the id and secret")
			flag.PrintDefaults()
			return
		}

		auth = &gddoexp.GithubAuth{
			ID:     *clientID,
			Secret: *clientSecret,
		}
	}

	// add cache to avoid repeated requests to Github
	gddoexp.HTTPClient = &http.Client{
		Transport: httpcache.NewTransport(
			diskcache.New(path.Join(os.Getenv("HOME"), ".gddoexp")),
		),
	}

	db, err := database.New()
	if err != nil {
		fmt.Println("error connecting to database:", err)
		return
	}

	pkgs, err := db.AllPackages()
	if err != nil {
		fmt.Println("error retrieving all packages:", err)
		return
	}

	file, err := os.OpenFile(*output, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("error creating output file:", err)
		return
	}
	defer file.Close()

	log.SetOutput(file)
	log.Println("BEGIN")
	log.Printf("%d packages will be analyzed", len(pkgs))

	var progressBar *pb.ProgressBar
	if progress != nil && *progress {
		progressBar = pb.StartNew(len(pkgs))
	}

	var cache int

	for response := range gddoexp.ShouldSuppressPackages(pkgs, db, auth) {
		if progress != nil && *progress {
			progressBar.Increment()
		}

		if response.Cache {
			cache++
		}

		if response.Error != nil {
			log.Println(response.Error)
		} else if response.Suppress {
			log.Printf("package “%s” should be suppressed\n", response.Package.Path)
			if progress != nil && !*progress {
				fmt.Println(response.Package.Path)
			}
		}
	}

	if progress != nil && *progress {
		progressBar.Finish()
	}

	log.Println("Cache hits:", cache)
	log.Println("END")
}
