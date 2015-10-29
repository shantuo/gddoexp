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
)

func main() {
	file := flag.String("file", "", "File containing the list of packages")
	output := flag.String("output", "gddoscore.out", "Output file")
	progress := flag.Bool("progress", false, "Show a progress bar")
	flag.Parse()

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

	db, err := database.New()
	if err != nil {
		fmt.Println("error connecting to database:", err)
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

	db.Do(func(pkg *database.PackageInfo) error {
		if _, ok := pkgs[pkg.PDoc.ImportPath]; !ok {
			// we aren't analyzing this package
			return nil
		}

		if pkg.Score == 0 {
			log.Printf("package “%s” has no score", pkg.PDoc.ImportPath)
		} else {
			log.Printf("package “%s” has score", pkg.PDoc.ImportPath)
			if progress != nil && !*progress {
				fmt.Println(pkg.PDoc.ImportPath)
			}
		}

		if progress != nil && *progress {
			progressBar.Increment()
		}

		return nil
	})

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

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file “%s”: %s\n", file, err)
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
