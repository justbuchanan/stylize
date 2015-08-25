from stylize.util import print_aligned, file_ext
from stylize.clang_formatter import ClangFormatter
from stylize.yapf_formatter import YapfFormatter

from itertools import chain
import argparse
import fcntl
import os
import struct
import subprocess
from multiprocessing.pool import ThreadPool
import sys
import termios


def enumerate_all_files(exclude=[]):
    for root, dirs, files in os.walk('.', topdown=True):
        dirs[:] = [d for d in dirs
                   if os.path.abspath(root + '/' + d) not in exclude]
        for f in files:
            yield root + '/' + f


## Yields all files that differ from @diffbase or are not tracked by git.
def enumerate_changed_files(exclude=[], diffbase="origin/master"):
    out = subprocess.check_output(
        "git diff --name-only %s; git ls-files --others --exclude-standard" %
        diffbase,
        shell=True)
    for line in out.decode("utf-8").splitlines():
        filepath = line.rstrip()
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
        help=
        "The git branch/tag/SHA1 to compare against.  If provided, only files that have changed since the diffbase will be scanned.")

    formatters = [ClangFormatter(), YapfFormatter()]

    # register any formatter-specific arguments
    formatters_by_ext = {}
    for formatter in formatters:
        if formatter.get_command() == None:
            print(
                "[ERR] A required dependency was not found. Check to see if clang-format is available on your path.")
            return 1
        formatter.add_args(parser)
        for ext in formatter.file_extensions:
            formatters_by_ext[ext] = formatter

    ARGS = parser.parse_args()

    ARGS.exclude_dirs = [os.path.abspath(p) for p in ARGS.exclude_dirs
                         ] + [os.path.abspath('.git')]

    # Print initial status info
    verb = "Checkstyling" if ARGS.check else "Formatting"
    if ARGS.diffbase:
        # A diffbase was given, so we run a git diff to see which files have
        # changed relative to the diffbase and only reformat those.  If a
        # formatter's config file has changed, we add all relevant files to the
        # list to format/checkstyle.

        print("%s files that differ from %s..." % (verb, ARGS.diffbase))

        changed_files = list(enumerate_changed_files(ARGS.exclude_dirs,
                                                     ARGS.diffbase))

        files_to_format = changed_files

        # Build a set of file extensions for which the config file has been modified.
        exts_requiring_full_reformat = set()
        for formatter in formatters:
            if formatter.config_file_name in changed_files:
                print(
                    "Config file '%s' changed.  %s all files with extensions: %s"
                    % (formatter.config_file_name, verb, str(
                        formatter.file_extensions)))
                exts_requiring_full_reformat |= set(formatter.file_extensions)

        if len(exts_requiring_full_reformat) > 0:
            files_with_relevant_extensions = filter(
                lambda file: file_ext(file) in exts_requiring_full_reformat,
                enumerate_all_files())
            files_to_format = chain(changed_files,
                                    files_with_relevant_extensions)
    else:
        print("%s all c++ and python files in the project..." % verb)
        files_to_format = enumerate_all_files(ARGS.exclude_dirs)

    def process_file(filepath):
        nonlocal file_scan_count
        nonlocal file_change_count
        nonlocal ARGS

        ext = file_ext(filepath)
        if ext not in formatters_by_ext:
            return
        formatter = formatters_by_ext[ext]

        needed_formatting = formatter.run(ARGS, filepath, ARGS.check)

        file_scan_count += 1
        if needed_formatting:
            file_change_count += 1

            suffix = "✗" if ARGS.check else "✔"
            print_aligned(filepath, suffix)
        else:
            print_aligned("> %s: %s" % (ext[1:], filepath),
                          "[%d]" % file_scan_count,
                          end="\r")

    # Use all the cores!
    workers = ThreadPool()
    workers.map(process_file, files_to_format)

    # Print final stats
    if ARGS.check:
        print_aligned(
            "[%d / %d] files need formatting" %
            (file_change_count, file_scan_count), "")
        return file_change_count
    else:
        print_aligned(
            "[%d / %d] files formatted" % (file_change_count, file_scan_count),
            "")
        return 0


if __name__ == '__main__':
    exit(main())
