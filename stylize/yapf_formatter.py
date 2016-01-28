from stylize.formatter import Formatter
from stylize.util import file_md5, bytes_md5

import subprocess
import shutil


class YapfFormatter(Formatter):
    def __init__(self):
        super().__init__(  )
        self.file_extensions= [      ".py"]
        self._config_file_name = ".style.yapf"

    def add_args(self, argparser):
        argparser.add_argument(
            "--yapf_style",
            type=str,
            default=None,
            help="The style to pass to yapf.  See `yapf --help` for more info")

    def run(self, args, filepath, check=False, calc_diff=False):
        logfile = open("/dev/null", "w")
        style_arg = "--style=%s" % (args.yapf_style if args.yapf_style != None
                                    else "pep8")
        if check or calc_diff:
            proc = subprocess.Popen(
                ["yapf", "--verify", "--diff", style_arg, filepath],
                stdout=subprocess.PIPE,
                stderr=logfile)
            out, err = proc.communicate()
            return out if len(out) > 0 else None, out.decode('utf-8')
        else:
            md5_before = file_md5(filepath)
            proc = subprocess.Popen(
                ["yapf", "-i", style_arg, filepath],
                stdout=logfile,
                stderr=logfile)
            proc.communicate()
            md5_after = file_md5(filepath)
            return (md5_before != md5_after), None

    def get_command(self):
        return shutil.which("yapf")
