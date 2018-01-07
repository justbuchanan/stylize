# Stylize [![circleci](https://circleci.com/gh/justbuchanan/stylize.svg?style=shield)](https://circleci.com/gh/justbuchanan/stylize) [![coveralls](https://coveralls.io/repos/justbuchanan/stylize/badge.svg?branch=master&service=github)](https://coveralls.io/github/justbuchanan/stylize?branch=master)


Stylize quickly reformats or checkstyles an entire repository of code.
It's a wrapper over other checkstyle programs such as `clang-format` or `yapf` that lets you use one command to operate on your entire repo, consisting of multiple types of files.


## Usage

~~~{.sh}
# install
go get -u gopkg.in/justbuchanan/stylize.v1

# format all code in-place (note: make a git commit before doing this - otherwise there's no undo button)
stylize.v1 -i

# format code in place, excluding the 'external' directory
stylize.v1 -i --exclude_dirs external

# generate a patch
stylize.v1 --patch_output patch.txt

# reformat files that differ from origin/master
stylize.v1 -i --git_diffbase origin/master
~~~


## Supported formatters

Stylize currently has support for:
* `buildifier`
* `clang-format`
* `gofmt`
* `yapf`

Other formatters can easily be added. See the \*\_formatter.go files as examples.


## Python version

This project is a rewrite of the original stylize, which was written in python.
Although it is no longer being developed, it's source code is available in the `python` branch of this repository.
