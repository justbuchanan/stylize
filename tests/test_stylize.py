from nose.tools import nottest
import os
import subprocess
import stylize.__main__ as stylize_main
import stylize.util as util
import tempfile
import unittest

BAD_CPP = b"int main() {\n\n\n\n}"
GOOD_CPP = b"int main() {}"
BAD_PY = b"a = 1+1"
GOOD_PY = b"a = 1 + 1\n"


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
        osenv = os.environ.copy()
        osenv["PYTHONPATH"] = os.path.dirname(__file__) + "/../"
        logfile = open(self.tempdir + "/test-log.txt", 'w')
        p = subprocess.Popen(cmd,
                             shell=True,
                             cwd=self.tempdir,
                             env=osenv,
                             stdout=logfile,
                             stderr=logfile)
        p.communicate()
        logfile.close()
        return p.returncode


## Add one bad cpp file and one good one, then ensure that only the bad one
# is reformatted.
class TestFormatCpp(Fixture):
    def test_format_cpp(self):
        self.write_file('bad.cpp', BAD_CPP)
        self.write_file('good.cpp', GOOD_CPP)

        self.assertNotEqual(0, self.run_cmd(
            "python3 -m stylize --clang_style=Google --check"))

        self.run_cmd("python3 -m stylize --clang_style=Google")

        self.assertEqual(0, self.run_cmd(
            "python3 -m stylize --clang_style=Google --check"))
        self.assertTrue(self.file_changed('bad.cpp', BAD_CPP))
        self.assertFalse(self.file_changed('good.cpp', GOOD_CPP))


## Add one bad py file and one good one, then ensure that only the bad one
# is reformatted.
class TestFormatPy(Fixture):
    def test_format_py(self):
        self.write_file('bad.py', BAD_PY)
        self.write_file('good.py', GOOD_PY)

        self.assertNotEqual(0, self.run_cmd("python3 -m stylize --check"))

        self.run_cmd("python3 -m stylize")

        self.assertEqual(0, self.run_cmd("python3 -m stylize --check"))
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

        self.run_cmd(
            "python3 -m stylize --clang_style=Google --diffbase=master")

        self.assertTrue(self.file_changed('bad2.cpp', BAD_CPP))
        self.assertFalse(self.file_changed('bad1.cpp', BAD_CPP))
