Stylize
=======

|image0| |image1|

Stylize is a command line interface for quickly reformatting a C++ or
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

.. |image0| image:: https://img.shields.io/pypi/v/stylize.svg
            :target: https://pypi.python.org/pypi/stylize
.. |image1| image:: https://img.shields.io/pypi/status/stylize.svg
            :target: https://pypi.python.org/pypi/stylize
