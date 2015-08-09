import fcntl
import hashlib
import struct
import sys
import termios


def file_md5(filepath):
    with open(filepath, 'r', encoding='utf-8') as f:
        try:
            return hashlib.md5(f.read().encode('utf-8', 'ignore')).hexdigest()
        except UnicodeDecodeError:
            print("ERROR encoding file: %s" % filepath)


def get_terminal_width():
    return struct.unpack('hh', fcntl.ioctl(sys.stdout, termios.TIOCGWINSZ,
                                           '1234'))[1]


## Print a left-justified string and a right-justified string by inserting the
# right amount of spaces in-between.
def print_justified(left, right, **print_args):
    spaces = " " * (get_terminal_width() - len(left) - len(right))
    print(left + spaces + right, **print_args)
