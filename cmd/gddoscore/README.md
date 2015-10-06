# gddoscore

gddoscore is a filter to identify only packages with non-zero scores. It can
work using an input file (one package per line) or via stdin.

For example, the following command would only print the packages with score that
should be archived:

```
% gddoexp -id cba321 -secret abc123 | gddoscore
```

You could also get some progress while running the filter, like the following
example:

```
% gddoscore -file gddoexp.txt -progress
303 / 3642 [=========>-------------------------------------------] 8.32 % 1m17s
```

When the progress bar isn't enabled only the packages with score are going to be
printed in the stdout. Otherwise, you could always check the output log, by
default is `gddoscore.out`.

For all options please check the `-h` flag:
```
% gddoscore -h
```