from setuptools import setup
import stylize

setup(name='stylize',
      packages=['stylize'],
      version=stylize.__version__,
      description=
      'A tool for quickly formatting and checkstyling C/C++ and Python code',
      long_description=open('readme.rst', 'r').read(),
      license='Apache License, Version 2.0',
      author='Justin Buchanan',
      author_email='justbuchanan@gmail.com',
      maintainer='Justin Buchanan',
      maintainer_email='justbuchanan@gmail.com',
      url='https://github.com/justbuchanan/stylize',
      classifiers=['Environment :: Console',
                   'Intended Audience :: Developers',
                   'Programming Language :: Python :: 3',
                   'Programming Language :: Python :: 3.4',
                   'Topic :: Software Development :: Quality Assurance', ],
      entry_points={'console_scripts': ['stylize = stylize.__main__:main'], },
      install_requires=['yapf==0.6.0'],
      test_suite='nose.collector',
      tests_require=['nose', 'coverage'], )
