# termcord

[![Tests](https://github.com/haguro/termcord/actions/workflows/tests.yml/badge.svg)](https://github.com/haguro/termcord/actions/workflows/tests.yml) [![Golint](https://github.com/haguro/termcord/actions/workflows/golint.yml/badge.svg)](https://github.com/haguro/termcord/actions/workflows/golint.yml) [![Go Report Card](https://goreportcard.com/badge/github.com/haguro/termcord)](https://goreportcard.com/report/github.com/haguro/termcord)

`termcord` is a terminal session recorder written in Go. It works by running a shell or a command in a pseudo-terminal and writing its output (optionally, its input as well) to a file. The aim for termcord is to match and expand upon the functionality of tools like `script`.

## Installation
Download the [latest release](https://github.com/haguro/termcord/releases/latest) for your system.

Or install with `go install`:
```
go install github.com/haguro/termcord@latest
```

## Features and usage
To start recording a shell session (session will be written to a file named by "termcording" by default):

```
$ termcord
```

To set the recording file name, use the `-f` flag followed by the desired filename. The following will start a shell session that will be recorded to a file named "foo":

```
$ termcord -f foo
```

To record the session of an arbitrary command execution instead, the command (and any of its arguments) can be passed to `termcord`. The following will execute `curl -I www.example.com` and record the execution output to the file "termcording":

```
$ termcord -curl -I www.example.com
```

To preview the recorded session:

```
$ cat termcording
```

For other optional flags and further usage information:

```
$ termcord -h
```
