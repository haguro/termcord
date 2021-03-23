# termcord

[![Tests](https://github.com/haguro/termcord/actions/workflows/tests.yml/badge.svg)](https://github.com/haguro/termcord/actions/workflows/tests.yml) [![Golint](https://github.com/haguro/termcord/actions/workflows/golint.yml/badge.svg)](https://github.com/haguro/termcord/actions/workflows/golint.yml) [![Go Report Card](https://goreportcard.com/badge/github.com/haguro/termcord)](https://goreportcard.com/report/github.com/haguro/termcord)

`termcord` is a terminal session recorder written in Go. It works by running a shell or a command in a pseudo-terminal and writes its output (and optionally its input) to a file. The aim is for termcord to match and expand upon functionality of tools like `script`.

## Installation
1. Install Go.
2. Run `go get -u github.com/haguro/termcord/cmd/termcord`.

## Features and usage
To start recording a shell session (session will be written to a file named by "tercording" by default):

```
$ termcord
```

To set the recording file name, we can pass it as an argument. The following will start a shell session recording that will written to a file named "foo":

```
$ termcord foo
```

To record a the execution session of an arbitrary command instead, the command (and any of its arguments can be passed ti `termcord` too). The following will execute `curl -I www.example.com` and records the execution output to a file named "foo":

```
$ termcord foo curl -I www.example.com
```

To preview a recorded session:

```
$ cat termcording
```

For optional flags and further usage information:

```
$ termcord -h
```
