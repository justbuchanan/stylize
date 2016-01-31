## Abstract superclass for all formatters. Each formatter defines a list of
# relevant file extensions and a run() method for doing the actual work.
class Formatter:
    def __init__(self):
        self.file_extensions = []
        self._config_file_name = None

    ## Adds any arguments to the given argparse.ArgumentParser object if needed.
    def add_args(self, argparser):
        pass

    ## The name of the config file used by this formatter - assumed to be at the
    # root of the project.
    @property
    def config_file_name(self):
        return self._config_file_name

    ## Run the formatter on the specified file.
    # @param args The arguments parsed by the ArgumentParser
    # @param filepath The file to check/format
    # @param check If true, run in checkstyle mode and don't modify the file.
    # @param calc_diff If true, the second return value of this function is the patch needed to bring the file into compliance
    # @return (noncompliant, patch) tuple.  @noncompliant is True if the file
    # needed/needs formatting and @patch contains a git-formatted patch if @calc_patch is True.
    def run(self, args, filepath, check=False, calc_diff=False):
        raise NotImplementedError("Subclass of Formatter must override run()")

    ## Checks if requirements are fullfilled and returns the command to use if they are
    # @return None if the required command is not found and the command to use by this formatter if found.
    def get_command(self):
        return None

    ## A list of file extensions that this formatter is relevant for.  Includes
    # the dot.
    @property
    def file_extensions(self):
        return self._file_extensions

    @file_extensions.setter
    def file_extensions(self, value):
        self._file_extensions = value
