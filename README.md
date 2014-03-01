para
====

Parallellizes stdin processing to a number of child processes.

Install
-------

Grab the binary from [the relases
page](https//github.com/calmh/para/releases).

Examples
--------

Perform line count in four parallell processes.

```
$ para -children=4 "wc -l" < /usr/share/dict/words
  150230
  152295
  150174
  151689
```

Perform a massively parallell grep. The `-quiet` flag supresses `Warning: exit status 1 (child exited with error)` for each of the `grep`s that exit without finding the search string.

```
$ para -children=64 -quiet "grep earde" < /usr/share/dict/words 
bearded
bearder
nonbearded
unbearded
```

Split a file into eight (the default number of children) pieces.
`$PARAIDX` is set to the child number for each child process.

```
$ para "cat > part.\$PARAIDX" < /usr/share/dict/words ; ls -l part.*
-rw-r--r--+ 1 jb  staff  525350 Mar  1 21:45 part.0
-rw-r--r--+ 1 jb  staff  536848 Mar  1 21:45 part.1
-rw-r--r--+ 1 jb  staff  585446 Mar  1 21:45 part.2
-rw-r--r--+ 1 jb  staff  513315 Mar  1 21:45 part.3
-rw-r--r--+ 1 jb  staff  527953 Mar  1 21:45 part.4
-rw-r--r--+ 1 jb  staff  520258 Mar  1 21:45 part.5
-rw-r--r--+ 1 jb  staff  536501 Mar  1 21:45 part.6
-rw-r--r--+ 1 jb  staff  547481 Mar  1 21:45 part.7

```

