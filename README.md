primcast
=======

[![Test Status](https://github.com/Songmu/primcast/workflows/test/badge.svg?branch=main)][actions]
[![Coverage Status](https://codecov.io/gh/Songmu/primcast/branch/main/graph/badge.svg)][codecov]
[![MIT License](https://img.shields.io/github/license/Songmu/primcast)][license]
[![PkgGoDev](https://pkg.go.dev/badge/github.com/Songmu/primcast)][PkgGoDev]

[actions]: https://github.com/Songmu/primcast/actions?workflow=test
[codecov]: https://codecov.io/gh/Songmu/primcast
[license]: https://github.com/Songmu/primcast/blob/main/LICENSE
[PkgGoDev]: https://pkg.go.dev/github.com/Songmu/primcast

The primcast is a primitive podcast site generator.

## Synopsis

```console
# Initialize the site
$ primcast init .

# Locate the audio file and create a new episode page
$ primcast episode audio/1.mp3
# episode/1.md is created

# Build the site
$ primcast build
# site generated under public/
```

## Description

The primcast is software that generates a minimum podcast sites from a list of audio files.

## Installation

```console
# Install the latest version. (Install it into ./bin/ by default).
% curl -sfL https://raw.githubusercontent.com/Songmu/primcast/main/install.sh | sh -s

# Specify installation directory ($(go env GOPATH)/bin/) and version.
% curl -sfL https://raw.githubusercontent.com/Songmu/primcast/main/install.sh | sh -s -- -b $(go env GOPATH)/bin [vX.Y.Z]

# In alpine linux (as it does not come with curl by default)
% wget -O - -q https://raw.githubusercontent.com/Songmu/primcast/main/install.sh | sh -s [vX.Y.Z]

# go install
% go install github.com/Songmu/primcast/cmd/primcast@latest
```

## Directory Structure

- index.md      - index page
- primcast.yaml - configuration file
- episode/      - episode pages in markdown
- audio/        - audio files (mp3 or m4a)
- template/     - template files (tmpl files in Go's text/template syntax)
- static/       - static files

## Sub Commmands

### init

```console
$ primcast init .
```

### episode

```
$ primcast episode [-slug=hoge -date=2024-09-01 -title=title] audio/1.mp3
```

create a new epoisode page with the specified audio file.

### build

```
$ primcast build
```

build the site and output to the `public` directory.

## Author

[Songmu](https://github.com/Songmu)
