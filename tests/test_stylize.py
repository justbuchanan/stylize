import unittest
import stylize.__main__ as stylize_main
from scripttest import TestFileEnvironment
import tempfile
import os




class TestStylize(unittest.TestCase):
    BAD_CPP=b"int main() {\n\n\n\n}"
    GOOD_CPP=b"int main() {}"
    BAD_PY=b"a = 1+1"
    GOOD_PY=b"a = 1 + 1\n"

    @classmethod
    def fresh_test_env(cls):
        osenv = os.environ.copy()
        osenv["PYTHONPATH"] = os.path.dirname(__file__) + "/../"
        env = TestFileEnvironment(tempfile.mkdtemp() + "/test", environ=osenv)
        return env

    ## Add one bad cpp file and one good one, then ensure that only the bad one
    # is reformatted
    def test_cpp_formatting(self):
        env = TestStylize.fresh_test_env()
        env.writefile('bad.cpp', TestStylize.BAD_CPP)
        env.writefile('good.cpp', TestStylize.GOOD_CPP)

        result = env.run("python3", "-m", "stylize", "--clang_style=file")

        self.assertTrue('bad.cpp' in result.files_updated)
        self.assertFalse('good.cpp' in result.files_updated)


    ## Add one bad py file and one good one, then ensure that only the bad one
    # is reformatted
    def test_py_formatting(self):
        env = TestStylize.fresh_test_env()
        env.writefile('bad.py', TestStylize.BAD_PY)
        env.writefile('good.py', TestStylize.GOOD_PY)

        result = env.run("python3", "-m", "stylize", "--clang_style=file")

        self.assertTrue('bad.py' in result.files_updated)
        self.assertFalse('good.py' in result.files_updated)

