import argparse
import fcntl
import struct
import sys
import termios

from stylize import stylize

print("module:DSFSDFDSFDSFDSFDS")
print(stylize)



num_so_far = num_changed = 0

def main():
    global num_so_far
    global num_changed

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



    TERM_WIDTH = struct.unpack('hh', fcntl.ioctl(sys.stdout, termios.TIOCGWINSZ,
                                             '1234'))[1]

    # Print initial status info
    verb = "Checkstyling" if ARGS.check else "Formatting"
    if ARGS.all:
        print("%s all c++ and python files in the project..." % verb)
        files_to_format = stylize.enumerate_all_files()
    else:
        print("%s files that differ from %s" % (verb, ARGS.diffbase))
        files_to_format = stylize.enumerate_changed_files(ARGS.diffbase)
    print("-" * TERM_WIDTH)





    def handle_file(filepath_and_type):
        filetype = filepath_and_type[0]
        filepath = filepath_and_type[1]

        if filetype == "cpp":
            needed_formatting = stylize.format_cpp(filepath, ARGS.check)
        elif filetype == "py":
            needed_formatting = stylize.format_py(filepath, ARGS.check)
        else:
            raise NameError("Unknown file type: %s" % filetype)

        global num_changed
        global num_so_far
        num_so_far += 1
        if needed_formatting:
            num_changed += 1

            suffix = "BAD" if ARGS.check else "FIXED"
            stylize.print_justified(filepath, suffix)
        else:
            stylize.print_justified("> %s: %s" % (filetype, filepath), "[%d]" % num_so_far,
                          end="\r")



    # Use all the cores!
    from multiprocessing.pool import ThreadPool
    workers = ThreadPool()
    workers.map(handle_file, stylize.get_files_to_format(files_to_format))

    # Print final stats
    if ARGS.check:
        stylize.print_justified(
            "[%d / %d] files need formatting" % (num_changed, num_so_far), "")
    else:
        stylize.print_justified("[%d / %d] files formatted" % (num_changed, num_so_far),
                        "")
