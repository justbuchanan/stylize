#!/usr/bin/env python3

import os
import sys
from subprocess import Popen
import multiprocessing
import subprocess
import struct
import fcntl
import termios

# logfile = open("reformat.log", "w")
logfile = open("/dev/null", "w")

TERM_WIDTH = struct.unpack('hh', fcntl.ioctl(sys.stdout, termios.TIOCGWINSZ,
                                             '1234'))[1]


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


def get_files_to_format(file_list):
    for filepath in file_list:
        _, ext = os.path.splitext(filepath)
        if ext in ['.c', '.cpp', '.h', '.hpp']:
            yield ("cpp", filepath)
        elif ext == '.py':
            yield ("py", filepath)
        else:
            continue


def print_justified(left, right, **print_args):
    spaces = " " * (TERM_WIDTH - len(left) - len(right))
    print(left + spaces + right, **print_args)
