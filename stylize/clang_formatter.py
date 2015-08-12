from stylize.formatter import Formatter
from stylize.util import file_md5

import os
import subprocess


class ClangFormatter(Formatter):
    def __init__(self):
        self.file_extensions = [".c", ".h", ".cpp", ".hpp"]

    def add_args(self, argparser):
        argparser.add_argument(
            "--clang_style",
            type=str,
            default=None,
            help=
            "The style to pass to clang-format.  See `clang-format --help` for more info.")

    def run(self, args, filepath, check=False):
        logfile = open("/dev/null", "w")
        if check:
            style_arg = "-style=%s" % args.clang_style if args.clang_style != None else ""
            return os.system(
                "clang-format %s -output-replacements-xml %s | grep '<replacement ' > /dev/null 2>&1"
                % (style_arg, filepath)) == 0
        else:
            md5_before = file_md5(filepath)
            popen_args = ["clang-format", "-i"]
            if args.clang_style:
                popen_args.append("-style=%s" % args.clang_style)
            popen_args.append(filepath)
            proc = subprocess.Popen(popen_args, stdout=logfile, stderr=logfile)
            proc.communicate()
            md5_after = file_md5(filepath)
            return md5_before != md5_after
