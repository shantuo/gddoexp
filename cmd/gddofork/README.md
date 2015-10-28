# gddofork

gddofork is a filter to remove forked packages created for small pull requests.
It can work using an input file (one package per line) or via stdin.

For example, the following command would only print the packages with score that
should be suppressed and aren't a fork for small pull requests:

```
% gddoexp -id cba321 -secret abc123 | gddoscore | gddofork -id cba321 -secret abc123
```

You could also get some progress while running the filter, like the following
example:

```
% gddofork -file packages.txt -progress
303 / 3642 [=========>-------------------------------------------] 8.32 % 1m17s
```

When the progress bar isn't enabled only the packages that aren't a fork for
small pull requests are going to be printed in the stdout. Otherwise, you could
always check the output log, by default is `gddofork.out`.

To run the program faster you should create a
[credential](https://github.com/settings/developers) in Github and pass it to
the program so we could get a more flexible rate limit.

For all options please check the `-h` flag:
```
% gddofork -h
```