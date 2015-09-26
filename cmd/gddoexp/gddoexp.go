// gddoexp is a command line tool crated to list eligible packages for
// archiving in GoDoc.org
package gddoexp

import (
	"fmt"

	"github.com/golang/gddo/database"
	"github.com/rafaeljusto/gddoexp"
)

func main() {
	db, err := database.New()
	if err != nil {
		fmt.Println("error creating database:", err)
		return
	}

	pkgs, err := db.AllPackages()
	if err != nil {
		fmt.Println("error retrieving all packages:", err)
		return
	}

	for _, pkg := range pkgs {
		if archive, err := gddoexp.ShouldArchivePackage(pkg.Path, db); err != nil {
			fmt.Println(err)
		} else if archive {
			fmt.Printf("package “%s” should be archived\n", pkg.Path)
		}
	}
}
