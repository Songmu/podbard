Podbard
=======

[![Test Status](https://github.com/Songmu/podbard/workflows/test/badge.svg?branch=main)][actions]
[![Coverage Status](https://codecov.io/gh/Songmu/podbard/branch/main/graph/badge.svg)][codecov]
[![MIT License](https://img.shields.io/github/license/Songmu/podbard)][license]
[![PkgGoDev](https://pkg.go.dev/badge/github.com/Songmu/podbard)][PkgGoDev]

[actions]: https://github.com/Songmu/podbard/actions?workflow=test
[codecov]: https://codecov.io/gh/Songmu/podbard
[license]: https://github.com/Songmu/podbard/blob/main/LICENSE
[PkgGoDev]: https://pkg.go.dev/github.com/Songmu/podbard

The Podbard is a primitive podcast site generator.

![](docs/static/images/artwork.jpg)

[Document site (Japanese)](https://junkyard.song.mu/podbard/)

## Synopsis

```console
# Initialize the site
$ podbard init .

# Locate the audio file and create a new episode page
$ podbard episode audio/1.mp3
# episode/1.md is created

# Build the site
$ podbard build
# site generated under public/
```

## Description

The podbard is software that generates a minimum podcast sites from a list of audio files.

## Installation

<details>
<summary>How to install on terminal</summary>

```console
# Install the latest version. (Install it into ./bin/ by default).
% curl -sfL https://raw.githubusercontent.com/Songmu/podbard/main/install.sh | sh -s

# Specify installation directory ($(go env GOPATH)/bin/) and version.
% curl -sfL https://raw.githubusercontent.com/Songmu/podbard/main/install.sh | sh -s -- -b $(go env GOPATH)/bin [vX.Y.Z]

# In alpine linux (as it does not come with curl by default)
% wget -O - -q https://raw.githubusercontent.com/Songmu/podbard/main/install.sh | sh -s [vX.Y.Z]

# go install
% go install github.com/Songmu/podbard/cmd/podbard@latest
```
</details>

## Directory Structure

- **index.md**
    - index page
- **podbard.yaml**
    - configuration file
- **episode/**
    - episode pages in markdown
- **audio/**
    - audio files (mp3 or m4a)
- **template/**
    - template files (tmpl files in Go's text/template syntax)
- **static/**
    - static files

## Sub Commmands

### init

```console
$ podbard init .
```

### episode

```
$ podbard episode [-slug=hoge -date=2024-09-01 -title=title] audio/1.mp3
```

create a new epoisode page with the specified audio file.

### build

```
$ podbard build
```

build the site and output to the `public` directory.

## Author

[Songmu](https://github.com/Songmu)
