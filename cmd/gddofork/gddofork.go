package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/cheggaaa/pb"
	"github.com/golang/gddo/database"
	"github.com/rafaeljusto/gddoexp"
)

func main() {
	clientID := flag.String("id", "", "Github client ID")
	clientSecret := flag.String("secret", "", "Github client secret")
	file := flag.String("file", "", "File containing the list of packages")
	output := flag.String("output", "gddofork.out", "Output file")
	progress := flag.Bool("progress", false, "Show a progress bar")
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

	var pkgs []database.Package
	var err error

	if file != nil && *file != "" {
		pkgs, err = readFromFile(*file)
	} else {
		pkgs, err = readFromStdin()
	}

	if err != nil {
		fmt.Println(err)
		return
	}

	o, err := os.OpenFile(*output, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("error creating output file:", err)
		return
	}
	defer o.Close()

	log.SetOutput(o)
	log.Println("BEGIN")
	log.Printf("%d packages will be analyzed\n", len(pkgs))

	var progressBar *pb.ProgressBar
	if progress != nil && *progress {
		progressBar = pb.StartNew(len(pkgs))
	}

	for response := range gddoexp.AreFastForkPackages(pkgs, auth) {
		if progress != nil && *progress {
			progressBar.Increment()
		}

		if response.Error != nil {
			log.Println(response.Error)
		} else if response.FastFork {
			log.Printf("package “%s” is a fast fork\n", response.Path)
			if progress != nil && !*progress {
				fmt.Println(response.Path)
			}
		}
	}

	if progress != nil && *progress {
		progressBar.Finish()
	}

	log.Println("END")
}

func readFromFile(file string) ([]database.Package, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("error opening file “%s”: %s\n", file, err)
	}
	defer f.Close()

	var pkgs []database.Package
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		pkgs = append(pkgs, database.Package{
			Path: scanner.Text(),
		})
	}
	return pkgs, nil
}

func readFromStdin() ([]database.Package, error) {
	var pkgs []database.Package
	reader := bufio.NewReader(os.Stdin)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("error reading from stdin: %s", err)
		}

		line = strings.TrimSpace(line)
		pkgs = append(pkgs, database.Package{
			Path: line,
		})
	}

	return pkgs, nil
}
