# gddoexp

[![GoDoc](https://godoc.org/github.com/rafaeljusto/gddoexp?status.svg)](https://godoc.org/github.com/rafaeljusto/gddoexp)

gddoexp (Go Doc Dot Org Expired) was created to indicate if a package from GoDoc
should be archived. Idea was born from [issue
320](https://github.com/golang/gddo/issues/320) of [gddo
project](https://github.com/golang/gddo) by [@garyburd](https://github.com/garyburd).
The following rules are current applied to verify if a package should be archived:

* No other packages reference the analyzed package
* Package wasn't modified in the last 2 years

## Install

```
go get -u github.com/rafaeljusto/gddoexp/...
```

Remember to add your `$GOPATH/bin` to your `$PATH`.

## Tools

Please check the specific documentation of each tool in the subdirectories.