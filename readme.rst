Stylize |pypi_version| |circleci_button| |coverage_button|
==========================================================

Stylize is a command line tool for quickly reformatting a C++ or
Python codebase. It’s a wrapper for `clang-format`_ and `yapf`_.

Install
-------

.. code:: sh

    pip3 install stylize

Usage
-----

.. code:: sh

    # reformat all C++/Python files in the current directory recursively
    stylize

    # reformat only files that differ from origin/master
    stylize --diffbase=origin/master

    # run in checkstyle mode - no files are changed and a nonzero return code
    # indicates that some files are out of accordance with the style configurations.
    stylize --check

.. _clang-format: http://clang.llvm.org/docs/ClangFormat.html
.. _yapf: https://github.com/google/yapf

.. |pypi_version| image:: https://img.shields.io/pypi/v/stylize.svg
            :target: https://pypi.python.org/pypi/stylize
.. |pypi_status| image:: https://img.shields.io/pypi/status/stylize.svg
            :target: https://pypi.python.org/pypi/stylize
.. |circleci_button| image:: https://circleci.com/gh/justbuchanan/stylize.svg?style=shield
            :target: https://circleci.com/gh/justbuchanan/stylize
.. |coverage_button| image:: https://coveralls.io/repos/justbuchanan/stylize/badge.svg?branch=master&service=github
  :target: https://coveralls.io/github/justbuchanan/stylize?branch=master
