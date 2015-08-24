from stylize.formatter import Formatter
from stylize.util import file_md5, bytes_md5

import subprocess
import shutil


class YapfFormatter(Formatter):
    def __init__(self):
        self.file_extensions = [".py"]
        self._config_file_name = ".style.yapf"

    def add_args(self, argparser):
        argparser.add_argument(
            "--yapf_style",
            type=str,
            default=None,
            help="The style to pass to yapf.  See `yapf --help` for more info")

    def run(self, args, filepath, check=False):
        logfile = open("/dev/null", "w")
        md5_before = file_md5(filepath)
        style_arg = "--style=%s" % (args.yapf_style if args.yapf_style != None
                                    else "pep8")
        if check:
            proc = subprocess.Popen(["yapf", "--verify", "--diff", style_arg,
                                     filepath],
                                    stdout=subprocess.PIPE,
                                    stderr=logfile)
            out, err = proc.communicate()
            return len(out) > 0
        else:
            proc = subprocess.Popen(["yapf", "-i", style_arg, filepath],
                                    stdout=logfile,
                                    stderr=logfile)
            proc.communicate()
            md5_after = file_md5(filepath)
            return md5_before != md5_after

    def get_command(self):
        return shutil.which("yapf")
