from nose.tools import nottest
from stylize.__main__ import main as stylize_main
import os
import stylize.util as util
import subprocess
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
        p = subprocess.Popen(
            cmd, shell=True, cwd=self.tempdir, stdout=logfile, stderr=logfile)
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

        self.assertNotEqual(0,
                            self.run_stylize(
                                ["--clang_style=Google", "--check"]))

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

        # commit bad file on master
        self.write_file('file1.cpp', BAD_CPP)
        self.run_cmd("git add .")
        self.run_cmd("git commit -m 'added poorly-formatted cpp file'")

        # modify file
        self.write_file('file1.cpp', (str(BAD_CPP) + '\n').encode('utf-8'))
        self.run_cmd("git add .")
        self.run_cmd("git commit -m 'modified file1.cpp'")

        # new branch off of first commit (not most recent on master)
        self.run_cmd("git checkout -b new-branch HEAD~1")

        # add a new bad file
        self.write_file('file2.cpp', BAD_CPP)

        # run stylize - the file shouldn't change b/c it was modified on master
        # *after* this branch was branched off
        self.run_stylize(["--clang_style=Google", "--diffbase=master"])
        self.assertFalse(self.file_changed('file1.cpp', BAD_CPP))
        self.assertTrue(self.file_changed('file2.cpp', BAD_CPP))

        # When the config file changes and we're using the --diffbase option,
        # all files with the extensions related to that config should be
        # formatted.
        self.write_file('.clang-format', EXAMPLE_CLANG_FORMAT)
        self.run_stylize(["--clang_style=file", "--diffbase=master"])
        self.assertTrue(self.file_changed('file1.cpp', BAD_CPP))


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

        self.run_stylize([
            "--clang_style=Google", "--diffbase=master", "--exclude_dirs",
            "dir1"
        ])
        self.assertFalse(self.file_changed('dir1/bad1.cpp', BAD_CPP))
        self.assertFalse(self.file_changed('dir1/bad2.cpp', BAD_CPP))


# Run stylize with an invalid --diffbase option
class TestInvalidDiffbase(Fixture):
    def test_invalid_diffbase(self):
        # commit file.cpp on master
        self.run_cmd('git init')
        self.write_file('file.cpp', BAD_CPP)
        self.run_cmd('git add .')
        self.run_cmd('git commit -m "commit"')

        # switch to a new branch and rebase the commit so 'master' and 'branch' have no history in common
        self.run_cmd('git checkout -b branch')
        self.run_cmd('git commit --amend -m "rebased commit"')

        # Run stylize with --diffbase=master, which is invalid since the current
        # branch has no commits in common with master.  Stylize should fallback
        # to doing a full reformat of the repository.
        self.run_stylize(["--clang_style=Google", "--diffbase=master"])
        self.assertTrue(self.file_changed('file.cpp', BAD_CPP))


## Test stylize's patch output feature
class TestPatchOutput(Fixture):
    def test_patch_output(self):
        # Init with both good and bad python and c++ files
        self.write_file("bad.cpp", BAD_CPP)
        self.write_file("good.cpp", GOOD_CPP)
        self.write_file("bad.py", BAD_PY)
        self.write_file("good.py", GOOD_PY)
        self.write_file(".gitignore", b"*.patch\n")

        # Setup git
        self.run_cmd("git init")
        self.run_cmd("git add --all")
        self.run_cmd("git commit -m 'first commit'")

        # Tell stylize to generate a patch file and check it
        self.run_stylize(
            ["--clang_style=Google", "--output_patch_file=pretty.patch"])
        self.assertTrue(os.path.isfile('pretty.patch'))

        # ensure that stylize didn't change any files (note that the patch file
        # is ignored by the .gitignore file)
        self.assertTrue(self.run_cmd("git diff --quiet") == 0)

        # ensure that applying the patch works without error
        self.assertTrue(self.run_cmd("git apply pretty.patch") == 0)

        # commit changes
        self.run_cmd("git add .")
        self.run_cmd("git commit -m 'clean'")

        # re-run stylize
        self.assertTrue(self.run_stylize(["--clang_style=Google"]) == 0)

        # ensure that stylize didn't need to change anything else
        self.assertTrue(self.run_cmd("git diff --quiet") == 0)
