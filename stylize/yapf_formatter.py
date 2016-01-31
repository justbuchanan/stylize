from stylize.formatter import Formatter
from stylize.util import *

import os
import shutil
import subprocess
import tempfile


class YapfFormatter(Formatter):
    def __init__(self):
        super().__init__()
        self.file_extensions = [".py"]
        self._config_file_name = ".style.yapf"
        self._tempdir = tempfile.mkdtemp()

    def add_args(self, argparser):
        argparser.add_argument(
            "--yapf_style",
            type=str,
            default=None,
            help="The style to pass to yapf.  See `yapf --help` for more info")

    def run(self, args, filepath, check=False, calc_diff=False):
        logfile = open("/dev/null", "w")
        style_arg = "--style=%s" % (args.yapf_style or "pep8")
        popen_args = ["yapf", style_arg, filepath]
        if check or calc_diff:
            # write style-compliant version of file to a tmp directory
            outfile_path = os.path.join(self._tempdir, filepath)
            os.makedirs(os.path.dirname(outfile_path), exist_ok=True)
            outfile = open(outfile_path, 'w')
            proc = subprocess.Popen(popen_args,
                                    stdout=outfile,
                                    stderr=subprocess.PIPE)
            out, err = proc.communicate()
            outfile.close()

            # return code zero indicates style-compliant file. 2 indicates non-
            # compliance.  Other return codes indicate errors.
            if proc.returncode != 0 and proc.returncode != 2:
                raise RuntimeError("Call to yapf failed for file '%s':\n%s" %
                                   (filepath, err.decode('utf-8')))

            # note: filepath[2:] cuts off leading './'
            patch = calculate_diff(filepath, outfile_path, filepath)
            noncompliant = len(patch) > 0

            return noncompliant, patch
        else:
            md5_before = file_md5(filepath)
            proc = subprocess.Popen(popen_args + ['-i'],
                                    stdout=logfile,
                                    stderr=logfile)
            proc.communicate()
            md5_after = file_md5(filepath)
            return (md5_before != md5_after), None

    def get_command(self):
        return shutil.which("yapf")
