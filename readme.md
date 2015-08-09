# Stylize
![](https://img.shields.io/pypi/v/stylize.svg) ![](https://img.shields.io/pypi/status/stylize.svg)

Stylize is a command line interface for quickly reformatting a C++ or Python codebase.  It's a wrapper for [clang-format](http://clang.llvm.org/docs/ClangFormat.html) and [yapf](https://github.com/google/yapf).


## Install

~~~{.sh}
pip3 install stylize
~~~


## Usage

~~~{.sh}
# reformat all C++/Python files in the current directory recursively
stylize

# reformat only files that differ from origin/master
stylize --diffbase=origin/master

# run in checkstyle mode - no files are changed and a nonzero return code
# indicates that some files are out of accordance with the style configurations.
stylize --check
~~~
