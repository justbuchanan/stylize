#!/usr/bin/env python3

import os
import sys
from subprocess import Popen
import multiprocessing
import subprocess

from stylize import util

# logfile = open("reformat.log", "w")
logfile = open("/dev/null", "w")

TERM_WIDTH = util.get_terminal_width()


def enumerate_all_files(exclude=[]):
    for root, dirs, files in os.walk('.', topdown=True):
        dirs[:] = [d for d in dirs if root + '/' + d not in exclude]
        for f in files:
            yield root + '/' + f


# TODO: ignore excluded dirs when git diffing
def enumerate_changed_files(exclude=[], diffbase="robojackets/master"):
    # TODO: respect the @exclude list
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
