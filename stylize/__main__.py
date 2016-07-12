from stylize.util import print_aligned, file_ext
from stylize.clang_formatter import ClangFormatter
from stylize.yapf_formatter import YapfFormatter
from stylize import __version__

from itertools import chain
from multiprocessing.pool import ThreadPool
import argparse
import fcntl
import os
import struct
import subprocess
import sys


def enumerate_all_files(exclude=[], directory='.'):
    for root, dirs, files in os.walk(directory, topdown=True):
        dirs[:] = [d
                   for d in dirs
                   if os.path.abspath(root + '/' + d) not in exclude]
        if root.startswith('./'): root = root[2:]
        for f in files:
            yield root + '/' + f


## Returns the checksum of the currently-checked-out commit
def current_git_commit():
    return subprocess.check_output(['git', 'rev-parse', 'HEAD']).strip(
    ).decode('utf-8')


## Yields all files that differ from the branching point with @diffbase or are
# not tracked by git.
def enumerate_changed_files(exclude=[], diffbase="origin/master"):
    # find common ancestor between @diffbase and @current_commit
    ancestor = subprocess.check_output(
        ['git', 'merge-base', current_git_commit(), diffbase]).strip().decode(
            'utf-8')

    # get list of files that have changed since @ancestor.
    out = subprocess.check_output([
        "git", "--no-pager", "diff", "--name-only", ancestor
    ])
    # Also include un-committed files.
    out += subprocess.check_output(['git', 'ls-files', '--others',
                                    '--exclude-standard'])

    # enumerate and yield relevant results
    for line in out.decode("utf-8").splitlines():
        filepath = line.rstrip()
        abspath = os.path.abspath(filepath)
        if os.path.exists(filepath):
            if not any(abspath.startswith(excluded_dir)
                       for excluded_dir in exclude):
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
        "Determine if all code is in accordance with the style configs, but don't fix them if they're not. A nonzero exit code indicates that some files don't meet the style requirements.")
    parser.add_argument("--exclude_dirs",
                        type=str,
                        default=[],
                        nargs="+",
                        help="A list of directories to exclude")
    parser.add_argument(
        "--output_patch_file",
        type=str,
        default=None,
        help=
        "If specified, a patch file is generated at the given path that, when aplied to the project, will fix all style mistakes.")
    parser.add_argument(
        "--diffbase",
        help=
        "The git branch/tag/SHA1 to compare against.  If provided, only files that have changed since the diffbase will be scanned.")
    parser.add_argument("--version",
                        action='store_true',
                        help="Print version and exit.")

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

    # version command
    if ARGS.version:
        print("stylize %s" % __version__)
        exit(0)

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
                    % (formatter.config_file_name, verb,
                       str(formatter.file_extensions)))
                exts_requiring_full_reformat |= set(formatter.file_extensions)

        if len(exts_requiring_full_reformat) > 0:
            files_with_relevant_extensions = filter(
                lambda file: file_ext(file) in exts_requiring_full_reformat,
                enumerate_all_files(ARGS.exclude_dirs))
            # use set() to eliminate any duplicates
            files_to_format = set(chain(changed_files,
                                        files_with_relevant_extensions))
    else:
        print("%s all c++ and python files in the project..." % verb)
        files_to_format = enumerate_all_files(ARGS.exclude_dirs)

    # This variable holds the final patch
    patch = ""

    def process_file(filepath):
        nonlocal patch
        nonlocal file_scan_count
        nonlocal file_change_count
        nonlocal ARGS

        ext = file_ext(filepath)
        if ext not in formatters_by_ext:
            return
        formatter = formatters_by_ext[ext]

        create_patch = ARGS.output_patch_file != None
        needed_formatting, patch_partial = formatter.run(
            ARGS, filepath, ARGS.check, create_patch)

        # concatenate all patches together
        if ARGS.output_patch_file and needed_formatting:
            patch += patch_partial + "\n"

        file_scan_count += 1
        if needed_formatting:
            file_change_count += 1

            status = "✗" if ARGS.check else "✔"
            print_aligned(filepath, status)
        else:
            print_aligned("> %s: %s" % (ext[1:], filepath),
                          "[%d]" % file_scan_count,
                          end="\r")

    # Use all the cores!
    workers = ThreadPool()
    workers.map(process_file, files_to_format)

    # Print final stats
    if ARGS.check:
        print_aligned("[%d / %d] files need formatting" %
                      (file_change_count, file_scan_count), "")
        retcode = 0 if file_change_count == 0 else 1
    else:
        print_aligned("[%d / %d] files formatted" %
                      (file_change_count, file_scan_count), "")
        retcode = 0

    if ARGS.output_patch_file:
        if file_change_count > 0:
            print("Writing patch to file: '%s'" % ARGS.output_patch_file)
            with open(ARGS.output_patch_file, 'w') as patchfile:
                patchfile.write(patch)
        else:
            print(
                "Skipping patch file generation, all files are style-compliant.")

    return retcode


if __name__ == '__main__':
    exit(main())
