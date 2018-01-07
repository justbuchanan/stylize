
# Stylize

Stylize quickly reformats or checkstyles an entire repository of code.
It's a wrapper over other checkstyle programs such as `clang-format` or `yapf`.


## Usage

~~~{.sh}
# install
go get -u gopkg.in/justbuchanan/stylize.v1

# format all code in-place
stylize.v1 -i

# format code in place, excluding the 'external' directory
stylize.v1 -i --exclude_dirs external

# generate a patch
stylize.v1 -o patch.txt

# reformat files that differ from origin/master
stylize.v1 -i --git_diffbase origin/master
~~~


## Supported formatters

Stylize currently has support for:
* `buildifier`
* `gofmt`
* `clang-format`
* `yapf`

Other formatters can easily be added. See the *_formatter.go files as examples.
