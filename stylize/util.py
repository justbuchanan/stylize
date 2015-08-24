import fcntl
import hashlib
import os
import struct
import sys
import termios


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


## Print a left-aligned string and a right-aligned string by inserting the
# right amount of spaces in-between.
def print_aligned(left, right, **print_args):
    spaces = " " * (get_terminal_width() - len(left) - len(right))
    print(left + spaces + right, **print_args)
