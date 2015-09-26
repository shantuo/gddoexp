package gddoexp_test

import (
	"fmt"
	"testing"

	"github.com/rafaeljusto/gddoexp"
)

func TestError(t *testing.T) {
	err := gddoexp.NewError("path/to/project", gddoexp.ErrorCodeRetrieveImportCounts, fmt.Errorf("i'm a crazy error"))
	expected := "[path/to/project] error retrieving import counts: i'm a crazy error"

	if msg := err.Error(); msg != expected {
		t.Errorf("expected “%s” and got “%s”", expected, msg)
	}

	err = gddoexp.NewError("path/to/project", gddoexp.ErrorCodeNonGithub, nil)
	expected = "[path/to/project] not a Github project"

	if msg := err.Error(); msg != expected {
		t.Errorf("expected “%s” and got “%s”", expected, msg)
	}
}
