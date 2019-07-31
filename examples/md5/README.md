MD5 Directory Digest
====================

This example uses Automi to print a list of MD5 digest for files in a given directory.  The
example uses `filepath.Walk` to visit subpaths and print the MD5 digest of each file. The
original example is from blog entry [Go Concurrency Patterns: Pipelines and Cancellation](https://blog.golang.org/pipelines).