package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/cheggaaa/pb"
	"github.com/golang/gddo/database"
)

// Given an input file with a list of packages (one per line), list all
// projects with score 0 (zero).

func main() {
	file := flag.String("file", "", "File containing the list of packages")
	output := flag.String("output", "score.out", "Output file")
	flag.Parse()

	if file == nil || *file == "" {
		fmt.Println("file not informed")
		flag.PrintDefaults()
		return
	}

	f, err := os.Open(*file)
	if err != nil {
		fmt.Printf("error opening file “%s”: %s\n", *file, err)
		return
	}
	defer f.Close()

	pkgs := make(map[string]bool)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		pkgs[scanner.Text()] = true
	}

	o, err := os.OpenFile(*output, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("error creating output file:", err)
		return
	}
	defer o.Close()

	db, err := database.New()
	if err != nil {
		fmt.Println("error connecting to database:", err)
		return
	}

	log.SetOutput(o)
	log.Println("BEGIN")
	log.Printf("%d packages will be analyzed\n", len(pkgs))

	progress := pb.StartNew(len(pkgs))
	db.Do(func(pkg *database.PackageInfo) error {
		if _, ok := pkgs[pkg.PDoc.ImportPath]; !ok {
			// we aren't analyzing this package
			return nil
		}

		if pkg.Score == 0 {
			log.Printf("package “%s” has no score", pkg.PDoc.ImportPath)
		} else {
			log.Printf("package “%s” has score", pkg.PDoc.ImportPath)
		}

		progress.Increment()
		return nil
	})

	progress.Finish()
	log.Println("END")
}
