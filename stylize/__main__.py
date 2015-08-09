from stylize import util
from stylize.clang_formatter import ClangFormatter
from stylize.yapf_formatter import YapfFormatter

import argparse
import fcntl
import struct
import subprocess
import sys
import termios
import os


def enumerate_all_files(exclude=[]):
    for root, dirs, files in os.walk('.', topdown=True):
        dirs[:] = [d for d in dirs
                   if os.path.abspath(root + '/' + d) not in exclude]
        for f in files:
            yield root + '/' + f


def enumerate_changed_files(exclude=[], diffbase="robojackets/master"):
    p = subprocess.Popen(["git", "diff", "--name-only", diffbase],
                         stdout=subprocess.PIPE)
    for line in p.stdout:
        filepath = line.rstrip().decode("utf-8")
        if os.path.exists(filepath):
            for excluded_dir in exclude:
                if filepath.startswith(excluded_dir):
                    continue
            yield filepath


def main():
    file_scan_count = file_change_count = 0

    # Command line options
    parser = argparse.ArgumentParser(
        description="Format and checkstyle C++ and Python code")
    parser.add_argument(
        "--check",
        action='store_true',
        help=
        "Determine if all code is in accordance with the style configs, but don't fix them if they're not. An nonzero exit code indicates that some files don't meet the style requirements.")
    parser.add_argument(
        "--exclude_dirs",
        type=str,
        default=[],
        nargs="+",
        help="A list of directories to exclude")
    parser.add_argument(
        "--diffbase",
        help="The git branch/tag/SHA1 to compare against.  If provided, only files that have changed since the diffbase will be scanned.")
    ARGS = parser.parse_args()

    ARGS.exclude_dirs = [os.path.abspath(p) for p in ARGS.exclude_dirs]

    # Print initial status info
    verb = "Checkstyling" if ARGS.check else "Formatting"
    if ARGS.diffbase:
        print("%s files that differ from %s..." % (verb, ARGS.diffbase))
        files_to_format = enumerate_changed_files(ARGS.exclude_dirs,
                                                  ARGS.diffbase)
    else:
        print("%s all c++ and python files in the project..." % verb)
        files_to_format = enumerate_all_files(ARGS.exclude_dirs)


    # map file extension to formatter
    formatters = [ClangFormatter(), YapfFormatter()]
    formatter_map = {}
    for f in formatters:
        for ext in f.file_extensions:
            formatter_map[ext] = f

    def process_file(filepath):
        nonlocal file_scan_count
        nonlocal file_change_count

        _, ext = os.path.splitext(filepath)
        if ext not in formatter_map:
            return
        formatter = formatter_map[ext]

        needed_formatting = formatter.run(filepath, ARGS.check)

        file_scan_count += 1
        if needed_formatting:
            file_change_count += 1

            suffix = "✗" if ARGS.check else "✔"
            util.print_justified(filepath, suffix)
        else:
            util.print_justified("> %s: %s" % (ext[1:], filepath),
                                 "[%d]" % file_scan_count, end="\r")

    # Use all the cores!
    from multiprocessing.pool import ThreadPool
    workers = ThreadPool()
    workers.map(process_file, files_to_format)

    # Print final stats
    if ARGS.check:
        util.print_justified(
            "[%d / %d] files need formatting" % (file_change_count, file_scan_count), "")
    else:
        util.print_justified(
            "[%d / %d] files formatted" % (file_change_count, file_scan_count), "")
