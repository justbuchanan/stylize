from stylize.formatter import Formatter
from stylize.util import file_md5

import subprocess


class ClangFormatter(Formatter):
    def __init__(self):
        self.file_extensions = [".c", ".h", ".cpp", ".hpp"]

    def run(self, filepath, check=False):
        logfile = open("/dev/null", "w")
        if check:
            return os.system(
                "clang-format -style=file -output-replacements-xml %s | grep '<replacement ' > /dev/null 2>&1"
                % filepath) == 0
        else:
            md5_before = file_md5(filepath)
            proc = subprocess.Popen(["clang-format", "-style=file", "-i",
                                     filepath],
                                    stdout=logfile,
                                    stderr=logfile)
            proc.communicate()
            md5_after = file_md5(filepath)
            return md5_before != md5_after
