
from setuptools import setup
import format_tool


setup(
  name='format-tool',
  version=format_tool.__version__,
  description='A tool for quickly formatting and checkstyling C/C++ and Python code',
  license='Apache License, Version 2.0',
  author='Justin Buchanan',
  maintainer='Justin Buchanan',
  maintainer_email='justbuchanan@gmail.com',
  classifiers=['Development Status :: 3 - Alpha',
                'Environment :: Console',
                'Intended Audience :: Developers',
                'Programming Language :: Python 3',
                'Programming Language :: Python 3.4',
                'Topic :: Software Development :: Libraries :: Python Modules',
                'Topic :: Software Development :: Quality Assurance',])
