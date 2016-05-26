import fcntl
import hashlib
import os
import struct
import sys
import termios
import subprocess


def file_md5(filepath):
    with open(filepath, 'r', encoding='utf-8') as f:
        try:
            return hashlib.md5(f.read().encode('utf-8', 'ignore')).hexdigest()
        except UnicodeDecodeError:
            print("ERROR encoding file: %s" % filepath)


def bytes_md5(bytes):
    return hashlib.md5(bytes).hexdigest()


def file_ext(filepath):
    _, ext = os.path.splitext(filepath)
    return ext


def get_terminal_width():
    try:
        return struct.unpack('hh', fcntl.ioctl(sys.stdout, termios.TIOCGWINSZ,
                                               '1234'))[1]
    except OSError as e:
        return 80


## Generates a git-compatible patch that when applied to @old_file, will result
#  in @new_file.  Note: Prepends 'a' and 'b' prefixes to the to/from file paths
#  for git compatibility.
# @param label The subpath to the file in the repository.
def calculate_diff(old_file, new_file, label):
    if not os.path.isfile(old_file):
        raise RuntimeError("Old file doesn't exist")
    if not os.path.isfile(new_file):
        raise RuntimeError("New file doesn't exist")
    if label.startswith('./'): label = label[2:]
    diffproc = subprocess.Popen(
        ['diff', '-Naur', old_file, new_file, '-L', 'a/%s' % label, '-L',
         'b/%s' % label],
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE)
    out, err = diffproc.communicate()
    if diffproc.returncode not in [0, 1]:
        raise RuntimeError(
            "Error calculating file diff, retcode=%d, err =\n%s",
            (diffproc.returncode, err))
    return out.decode('utf-8')


## Print a left-aligned string and a right-aligned string by inserting the
# right amount of spaces in-between.
def print_aligned(left, right, **print_args):
    spaces = " " * (get_terminal_width() - len(left) - len(right))
    print(left + spaces + right, **print_args)
