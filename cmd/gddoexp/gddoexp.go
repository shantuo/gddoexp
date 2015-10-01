package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/golang/gddo/database"
	"github.com/gregjones/httpcache"
	"github.com/gregjones/httpcache/diskcache"
	"github.com/rafaeljusto/gddoexp"
)

func main() {
	clientID := flag.String("id", "", "Github client ID")
	clientSecret := flag.String("secret", "", "Github client secret")
	flag.Parse()

	var auth *gddoexp.GithubAuth
	if clientID != nil || clientSecret != nil {
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

	fmt.Printf("%d packages will be analyzed\n", len(pkgs))

	for response := range gddoexp.ShouldArchivePackages(pkgs, db, auth) {
		if response.Error != nil {
			fmt.Println(err)
		} else if response.Archive {
			fmt.Printf("package “%s” should be archived\n", response.Path)
		}
	}
}
