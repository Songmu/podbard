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

primcast short description

## Synopsis

```go
// simple usage here
```

## Description

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

## Author

[Songmu](https://github.com/Songmu)
