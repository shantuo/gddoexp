# gddoexp

[![GoDoc](https://godoc.org/github.com/rafaeljusto/gddoexp?status.svg)](https://godoc.org/github.com/rafaeljusto/gddoexp) [![Build Status](https://travis-ci.org/rafaeljusto/gddoexp.svg)](https://travis-ci.org/rafaeljusto/gddoexp) [![Coverage Status](https://coveralls.io/repos/rafaeljusto/gddoexp/badge.svg?branch=master&service=github)](https://coveralls.io/github/rafaeljusto/gddoexp?branch=master)

gddoexp (Go Doc Dot Org Expired) was created to indicate if a package from GoDoc
should be suppressed from search results. Idea was born from [issue
320](https://github.com/golang/gddo/issues/320) of [gddo
project](https://github.com/golang/gddo) by [@garyburd](https://github.com/garyburd).
The following rules are current applied to verify if a package should be
suppressed:

* No other packages reference the analyzed package
* Package wasn't modified in the last 2 years
* Package is a fork with a few commits (fast fork)

A fast fork package is a fork created to made some small changes for a pull
request. Currently we tolerate up to 2 commits in a period of 1 week after the
fork date.

## Install

```
go get -u github.com/rafaeljusto/gddoexp/...
```

Remember to add your `$GOPATH/bin` to your `$PATH`.

## Tools

Please check the specific documentation of each tool in the subdirectories.
