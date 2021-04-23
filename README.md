[![pkg.go.dev](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/korylprince/go-fs-walk)

# About

`go-fs-walk` provides a `Cursor` that recursively walks through filesystems (including archive files like zip and tgz). Filters can be used to control visited files and folders.

If you need to easily access files inside zips that are inside three layers of tars, this is the package for you.

# Installing

Using Go Modules:

`go get github.com/korylprince/go-fs-walk`

If you have any issues or questions [create an issue](https://github.com/korylprince/go-fs-walk/issues).

# Usage

See example on [pkg.go.dev](https://pkg.go.dev/github.com/korylprince/go-fs-walk?tab=doc#pkg-examples).

# To Do

* Add more archive formats (Send a PR!)
* Add test data
* Add safety for untrusted inputs (see below)

# Safety

`Cursor` is neither thread-safe, nor should it be used on untrusted data.
