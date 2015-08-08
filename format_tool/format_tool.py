#!/usr/bin/env python3

import os
import sys
from subprocess import Popen
import multiprocessing
import subprocess
import struct
import fcntl
import termios
import argparse


# yapf: disable
exclude_directories = set([
    './build',
    './third_party',
    './firmware/build',
    './firmware/robot/cpu/at91sam7s256',
    './firmware/robot/cpu/at91sam7s321',
    './firmware/robot/cpu/at91sam7s64'
])
# yapf: enable


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
    "--diffbase",
    default="robojackets/master",
    help="The branch/tag/SHA1 in git to compare against.")
parser.add_argument("--exclude_dirs", nargs="*")
ARGS = parser.parse_args()

# logfile = open("reformat.log", "w")
logfile = open("/dev/null", "w")


def file_md5(filepath):
    with open(filepath, 'r', encoding='utf-8') as f:
        import hashlib
        try:
            return hashlib.md5(f.read().encode('utf-8', 'ignore')).hexdigest()
        except UnicodeDecodeError:
            print("ERROR encoding file: %s" % filepath)


def format_py(filename, check=False):
    if check:
        proc = subprocess.Popen(["yapf", "--verify", filename],
                                stdout=logfile,
                                stderr=logfile)
        proc.communicate()
        return proc.returncode == 1
    else:
        md5_before = file_md5(filename)
        proc = subprocess.Popen(["yapf", "-i", filename],
                                stdout=logfile,
                                stderr=logfile)
        proc.communicate()
        md5_after = file_md5(filename)
        return md5_before != md5_after


def format_cpp(filename, check=False):
    if check:
        return os.system(
            "clang-format -style=file -output-replacements-xml %s | grep '<replacement ' > /dev/null 2>&1"
            % filename) == 0
    else:
        md5_before = file_md5(filename)
        proc = subprocess.Popen(["clang-format", "-style=file", "-i", filename
                                 ],
                                stdout=logfile,
                                stderr=logfile)
        proc.communicate()
        md5_after = file_md5(filename)
        return md5_before != md5_after


def enumerate_all_files():
    for root, dirs, files in os.walk('.', topdown=True):
        dirs[:] = [d for d in dirs if root + '/' + d not in exclude_directories
                   ]
        for f in files:
            yield root + '/' + f


# TODO: ignore excluded dirs when git diffing
def enumerate_changed_files(diffbase="robojackets/master"):
    p = subprocess.Popen(["git", "diff", "--name-only", diffbase],
                         stdout=subprocess.PIPE)
    for line in p.stdout:
        yield line.rstrip().decode("utf-8")


TERM_WIDTH = struct.unpack('hh', fcntl.ioctl(sys.stdout, termios.TIOCGWINSZ,
                                             '1234'))[1]

# Print initial status info
verb = "Checkstyling" if ARGS.check else "Formatting"
if ARGS.all:
    print("%s all c++ and python files in the project..." % verb)
    files_to_format = enumerate_all_files()
else:
    print("%s files that differ from %s" % (verb, ARGS.diffbase))
    files_to_format = enumerate_changed_files(ARGS.diffbase)
print("-" * TERM_WIDTH)


def get_files_to_format():
    for filepath in files_to_format:
        _, ext = os.path.splitext(filepath)
        if ext in ['.c', '.cpp', '.h', '.hpp']:
            yield ("cpp", filepath)
        elif ext == '.py':
            yield ("py", filepath)
        else:
            continue


num_so_far = num_changed = 0


def print_justified(left, right, **print_args):
    spaces = " " * (TERM_WIDTH - len(left) - len(right))
    print(left + spaces + right, **print_args)


def handle_file(filepath_and_type):
    filetype = filepath_and_type[0]
    filepath = filepath_and_type[1]

    if filetype == "cpp":
        needed_formatting = format_cpp(filepath, ARGS.check)
    elif filetype == "py":
        needed_formatting = format_py(filepath, ARGS.check)
    else:
        raise NameError("Unknown file type: %s" % filetype)

    global num_changed
    global num_so_far
    num_so_far += 1
    if needed_formatting:
        num_changed += 1

        suffix = "BAD" if ARGS.check else "FIXED"
        print_justified(filepath, suffix)
    else:
        print_justified("> %s: %s" % (filetype, filepath), "[%d]" % num_so_far,
                        end="\r")

# Use all the cores!
from multiprocessing.pool import ThreadPool
workers = ThreadPool()
workers.map(handle_file, get_files_to_format())

# Print final stats
if ARGS.check:
    print_justified(
        "[%d / %d] files need formatting" % (num_changed, num_so_far), "")
else:
    print_justified("[%d / %d] files formatted" % (num_changed, num_so_far),
                    "")
