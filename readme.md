
# Stylize

Stylize is a simple program for formatting or style-checking an entire repository of code. It runs checkstyle programs such as `clang-format` or `yapf` on your code.


## Usage

~~~{.sh}
# install
go get -u gopkg.in/justbuchanan/stylize.v1

# generate a patch
stylize.v1 -o patch.txt

# format code in place
stylize.v1 -i --exclude_dirs external
~~~
