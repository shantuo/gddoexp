# gddoexp

[![GoDoc](https://godoc.org/github.com/rafaeljusto/gddoexp?status.svg)](https://godoc.org/github.com/rafaeljusto/gddoexp)

gddoexp (Go Doc Dot Org Expired) was created to indicate if a package from GoDoc
should be archived. Idea was born from [issue
320](https://github.com/golang/gddo/issues/320) of [gddo
project](https://github.com/golang/gddo) by @garyburd. The following rules are
current applied to verify if a package should be archived:

* No other packages reference the analyzed package
* Package wasn't modified is the last 2 years

## Install

```
go get -u github.com/rafaeljusto/gddoexp/...
```

Remember to add your `$GOPATH/bin` to your `$PATH`.

## Command line usage

```
% gddoexp -h
Usage of ./gddoexp:
  -db-idle-timeout duration
    	Close Redis connections after remaining idle for this duration. (default 4m10s)
  -db-log
    	Log database commands
  -db-server string
    	URI of Redis server. (default "redis://127.0.0.1:6379")
  -id string
    	Github client ID
  -secret string
    	Github client secret
```
