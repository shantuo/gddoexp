# gddoexp

gddoexp is the command line that uses the library.

By default, the tool will connect to your local Redis server to retrieve the
gddo database, you could change that via parameters.

To run the program faster you should create a
[credential](https://github.com/settings/developers) in Github and pass it to
the program so we could get a more flexible rate limit.

You could also get some progress while running the tool, like the following
example:

```
% gddoexp -id cba321 -secret abc123 -progress
15 / 132277 [>----------------------------------------------] 0.01 % 12h45m10s
```

When the progress bar isn't enabled only the packages selected to be archived
are going to be printed in the stdout. Otherwise, you could always check the
output log, by default is `gddoexp.out`.

This tool contains a local cache for the Github responses that will be stored in
`$HOME/.gddoexp`. This is useful to avoid repeated queries to Github API.

For all options please check the `-h` flag:
```
% gddoexp -h
```