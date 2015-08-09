import argparse
import fcntl
import struct
import subprocess
import sys
import termios
import os

from stylize import util

from stylize.clang_formatter import ClangFormatter
from stylize.yapf_formatter import YapfFormatter


def enumerate_all_files(exclude=[]):
    for root, dirs, files in os.walk('.', topdown=True):
        dirs[:] = [d for d in dirs
                   if os.path.abspath(root + '/' + d) not in exclude]
        for f in files:
            yield root + '/' + f


# TODO: ignore excluded dirs when git diffing
def enumerate_changed_files(exclude=[], diffbase="robojackets/master"):
    # TODO: respect the @exclude list
    p = subprocess.Popen(["git", "diff", "--name-only", diffbase],
                         stdout=subprocess.PIPE)
    for line in p.stdout:
        filepath = line.rstrip().decode("utf-8")
        if os.path.exists(filepath):
            yield filepath


num_so_far = num_changed = 0


def main():
    global num_so_far
    global num_changed

    # Command line options
    parser = argparse.ArgumentParser(
        description="Format and checkstyle C++ and Python code")
    parser.add_argument(
        "--check",
        action='store_true',
        help=
        "Determine if all code is in accordance with the style configs, but don't fix them if they're not")
    parser.add_argument(
        "--all",
        action="store_true",
        help=
        "By default, we only format or checkstyle files that differ from the diffbase.  Pass --all to instead check all files in the repo")
    parser.add_argument(
        "--exclude_dirs",
        type=str,
        default=[],
        nargs="+",
        help="A list of directories to exclude")
    parser.add_argument(
        "--diffbase",
        default="origin/master",
        help="The branch/tag/SHA1 in git to compare against.")
    ARGS = parser.parse_args()

    ARGS.exclude_dirs = [os.path.abspath(p) for p in ARGS.exclude_dirs]

    # Print initial status info
    verb = "Checkstyling" if ARGS.check else "Formatting"
    if ARGS.all:
        print("%s all c++ and python files in the project..." % verb)
        files_to_format = enumerate_all_files(ARGS.exclude_dirs)
    else:
        print("%s files that differ from %s..." % (verb, ARGS.diffbase))
        files_to_format = enumerate_changed_files(ARGS.exclude_dirs,
                                                  ARGS.diffbase)
    # print("-" * util.get_terminal_width())

    formatters = [ClangFormatter(), YapfFormatter()]

    # map file extension to formatter
    formatter_map = {}
    for f in formatters:
        for ext in f.file_extensions:
            formatter_map[ext] = f

    def handle_file(filepath):
        global num_so_far
        global num_changed

        _, ext = os.path.splitext(filepath)
        if ext not in formatter_map:
            return
        formatter = formatter_map[ext]

        needed_formatting = formatter.run(filepath, ARGS.check)

        num_so_far += 1
        if needed_formatting:
            num_changed += 1

            suffix = "BAD" if ARGS.check else "FIXED"
            util.print_justified(filepath, suffix)
        else:
            util.print_justified("> %s: %s" % (ext[1:], filepath),
                                 "[%d]" % num_so_far,
                                 end="\r")

    # Use all the cores!
    from multiprocessing.pool import ThreadPool
    workers = ThreadPool()
    workers.map(handle_file, files_to_format)

    # Print final stats
    if ARGS.check:
        util.print_justified(
            "[%d / %d] files need formatting" % (num_changed, num_so_far), "")
    else:
        util.print_justified(
            "[%d / %d] files formatted" % (num_changed, num_so_far), "")
