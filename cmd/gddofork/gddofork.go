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

	var pkgs map[string]bool
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

	// TODO: We need to query the repository in Github [1] to see if it's a
	// fork (forked) and than check the commits [2] to see if they were all
	// made near the forked date (created_at). We want to detect projects
	// created only for small pull requests.
	//
	// [1] https://developer.github.com/v3/repos/#get
	// [2] https://developer.github.com/v3/repos/commits/#list-commits-on-a-repository

	if progress != nil && *progress {
		progressBar.Finish()
	}

	log.Println("END")
}

func readFromFile(file string) (map[string]bool, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("error opening file “%s”: %s\n", file, err)
	}
	defer f.Close()

	pkgs := make(map[string]bool)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		pkgs[scanner.Text()] = true
	}
	return pkgs, nil
}

func readFromStdin() (map[string]bool, error) {
	pkgs := make(map[string]bool)
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
		pkgs[line] = true
	}

	return pkgs, nil
}
