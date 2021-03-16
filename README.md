# termcord

[![Tests](https://github.com/haguro/termcord/actions/workflows/tests.yml/badge.svg)](https://github.com/haguro/termcord/actions/workflows/tests.yml) [![Golint](https://github.com/haguro/termcord/actions/workflows/golint.yml/badge.svg)](https://github.com/haguro/termcord/actions/workflows/golint.yml) 

`termcord` is a terminal session recorder written in Go.

## Features and usage
`termcord` is still a *work in progress* and currently only supports minimal functionality.

To start a recording session (session will be written to a file named _termcording_)

```
$ termcord
```

To preview a recorded session:

```
$ cat termcording
```

For usage information:

```
$ termcording -h
```

The aim is for termcord to match and expand upon functionality of tools like `script`.
