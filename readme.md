# Stylize [![circleci](https://circleci.com/gh/justbuchanan/stylize.svg?style=shield)](https://circleci.com/gh/justbuchanan/stylize) [![coveralls](https://coveralls.io/repos/justbuchanan/stylize/badge.svg?branch=master&service=github)](https://coveralls.io/github/justbuchanan/stylize?branch=master)

Stylize quickly reformats or checkstyles an entire repository of code.
It's a wrapper over other checkstyle programs such as `clang-format` or `yapf` that lets you use one command to operate on your entire repo, consisting of multiple types of files.

## Usage

```.sh
# install
go get -u github.com/justbuchanan/stylize

# check files and write a patch file to 'patch.txt'. This patch file shows what
# changes the formatter would have made if run with the `-i` (in-place) flag.
# You can also apply this generated patch to the repo using `git apply`.
stylize --patch_output patch.txt

# format all code in-place
# note: make a git commit before doing this - there's no undo button
stylize -i

# format code in place, excluding a couple directories
stylize -i --exclude=build,external

# reformat only files that differ from origin/master
stylize -i --git_diffbase origin/master
```

## Configuration

By default, `stylize` looks for a config file named `.stylize.yml` in the current directory. A different file can be specified with the `--config` flag. See [`config.go`](config.go) for what options are available and see this repo's [`.stylize.yml`](.stylize.yml) file as an example.

## Supported formatters

Stylize currently has support for:

-   [buildifier](https://github.com/bazelbuild/buildtools/blob/master/buildifier/README.md)
-   [clang-format](https://clang.llvm.org/docs/ClangFormat.html)
-   [gofmt](https://golang.org/cmd/gofmt/)
-   [yapf](https://github.com/google/yapf)
-   [prettier](https://github.com/prettier/prettier)
-   [uncrustify](https://github.com/uncrustify/uncrustify)
-   [rustfmt](https://github.com/rust-lang-nursery/rustfmt)

Other formatters can easily be added. See the files in the 'formatters' directory as examples.

## Python version

This project is a rewrite of the original stylize, which was written in python.
Although it is no longer being developed, it's source code is available in the `python` branch of this repository.
