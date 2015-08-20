from stylize.formatter import Formatter
from stylize.util import file_md5, bytes_md5

import subprocess
import shutil


class YapfFormatter(Formatter):
    def __init__(self):
        self.file_extensions = [".py"]

    def run(self, args, filepath, check=False):
        logfile = open("/dev/null", "w")
        md5_before = file_md5(filepath)
        if check:
            proc = subprocess.Popen(["yapf", "--verify", filepath],
                                    stdout=subprocess.PIPE,
                                    stderr=logfile)
            out, err = proc.communicate()
            return md5_before != bytes_md5(out)
        else:
            proc = subprocess.Popen(["yapf", "-i", filepath],
                                    stdout=logfile,
                                    stderr=logfile)
            proc.communicate()
            md5_after = file_md5(filepath)
            return md5_before != md5_after

    def get_command(self):
        return shutil.which("yapf")
