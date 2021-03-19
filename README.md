# termcord

[![Tests](https://github.com/haguro/termcord/actions/workflows/tests.yml/badge.svg)](https://github.com/haguro/termcord/actions/workflows/tests.yml) [![Golint](https://github.com/haguro/termcord/actions/workflows/golint.yml/badge.svg)](https://github.com/haguro/termcord/actions/workflows/golint.yml) [![Go Report Card](https://goreportcard.com/badge/github.com/haguro/termcord)](https://goreportcard.com/report/github.com/haguro/termcord)

`termcord` is a terminal session recorder written in Go.

## Installation
1. Install Go.
2. Run `go get -u github.com/haguro/termcord/cmd/termcord`.

## Features and usage
`termcord` is still a *work in progress* and currently only supports minimal functionality.

To start a recording session (session will be written to a file named "termcording")

```
$ termcord
```

To preview a recorded session:

```
$ cat termcording
```

For usage information:

```
$ termcord -h
```

The aim is for termcord to match and expand upon functionality of tools like `script`.
