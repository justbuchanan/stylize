from nose.tools import nottest
import os
import subprocess
from stylize.__main__ import main as stylize_main
import stylize.util as util
import sys
import tempfile
import unittest

BAD_CPP = b"int main() {\n\n\n\n}"
GOOD_CPP = b"int main() {}"
BAD_PY = b"a = 1+1"
GOOD_PY = b"a = 1 + 1\n"
EXAMPLE_CLANG_FORMAT = b"---\nBasedOnStyle: Google"


## Test fixture that sets up a temporary directory and provides some basic
# methods that we need for our tests.
class Fixture(unittest.TestCase):
    def __init__(self, *args, **kwargs):
        self.tempdir = tempfile.mkdtemp()
        super(Fixture, self).__init__(*args, **kwargs)

    @nottest
    def file_changed(self, filename, prev_contents):
        filepath = self.tempdir + "/" + filename
        return util.file_md5(filepath) != util.bytes_md5(prev_contents)

    @nottest
    def write_file(self, filename, contents):
        with open(self.tempdir + "/" + filename, 'wb') as f:
            f.write(contents)

    @nottest
    def run_cmd(self, cmd):
        logfile = open(self.tempdir + "/test-log.txt", 'w')
        p = subprocess.Popen(cmd,
                             shell=True,
                             cwd=self.tempdir,
                             stdout=logfile,
                             stderr=logfile)
        p.communicate()
        logfile.close()
        return p.returncode

    @nottest
    def run_stylize(self, args=[]):
        os.chdir(self.tempdir)
        sys.argv = ["stylize"] + args
        return stylize_main()


## Add one bad cpp file and one good one, then ensure that only the bad one
# is reformatted.
class TestFormatCpp(Fixture):
    def test_format_cpp(self):
        self.write_file('bad.cpp', BAD_CPP)
        self.write_file('good.cpp', GOOD_CPP)

        self.assertNotEqual(
            0, self.run_stylize(["--clang_style=Google", "--check"]))

        self.assertEqual(0, self.run_stylize(["--clang_style=Google"]))
        self.assertTrue(self.file_changed('bad.cpp', BAD_CPP))
        self.assertFalse(self.file_changed('good.cpp', GOOD_CPP))


## Add one bad py file and one good one, then ensure that only the bad one
# is reformatted.
class TestFormatPy(Fixture):
    def test_format_py(self):
        self.write_file('bad.py', BAD_PY)
        self.write_file('good.py', GOOD_PY)

        self.assertNotEqual(0, self.run_stylize(["--check"]))

        self.run_stylize()

        self.assertEqual(0, self.run_stylize(["--check"]))
        self.assertTrue(self.file_changed('bad.py', BAD_PY))
        self.assertFalse(self.file_changed('good.py', GOOD_PY))


## Commit a bad cpp file to the master branch, then add another bad one.
# Ensure that the committed one is not reformatted when we give stylize
# the --diffbase=master option.
class TestDiffbase(Fixture):
    def test_diffbase(self):
        self.run_cmd("git init")
        self.write_file('bad1.cpp', BAD_CPP)
        self.run_cmd("git add bad1.cpp")
        self.run_cmd("git commit -m 'added poorly-formatted cpp file'")
        self.write_file('bad2.cpp', BAD_CPP)

        self.run_stylize(["--clang_style=Google", "--diffbase=master"])

        self.assertTrue(self.file_changed('bad2.cpp', BAD_CPP))
        self.assertFalse(self.file_changed('bad1.cpp', BAD_CPP))

        # When the config file changes and we're using the --diffbase option,
        # all files with the extensions related to that config should be
        # formatted.
        self.write_file('.clang-format', EXAMPLE_CLANG_FORMAT)
        self.run_stylize(["--clang_style=file", "--diffbase=master"])
        self.assertTrue(self.file_changed('bad1.cpp', BAD_CPP))


## Test to ensure that stylize respects the "--exclude_dirs" option when it's
# also given the --diffbase option.
class TestDiffbaseExclude(Fixture):
    def test_diffbase_exclude(self):
        self.run_cmd("git init")
        self.run_cmd("mkdir dir1")
        self.write_file('dir1/bad1.cpp', BAD_CPP)
        self.run_cmd("git add dir1/bad1.cpp")
        self.run_cmd("git commit -m 'added poorly-formatted cpp file'")
        self.write_file('dir1/bad2.cpp', BAD_CPP)

        self.run_stylize(["--clang_style=Google", "--diffbase=master",
                          "--exclude_dirs", "dir1"])
        self.assertFalse(self.file_changed('dir1/bad1.cpp', BAD_CPP))
        self.assertFalse(self.file_changed('dir1/bad2.cpp', BAD_CPP))
