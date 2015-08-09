## Abstract superclass for all formatters. Each formatter defines a list of
# relevant file extensions and a run() method for doing the actual work.
class Formatter:
    def __init__(self):
        self.file_extensions = []

    ## Run the formatter on the specified file.
    # @param check If true, run in checkstyle mode and don't modify the file.
    # @return True if the file needed/needs formatting
    def run(self, filepath, check=False):
        raise NotImplementedError("Subclass of Formatter must override run()")

    ## A list of file extensions that this formatter is relevant for.  Included
    # the dot.
    @property
    def file_extensions(self):
        return self._file_extensions
    @file_extensions.setter
    def file_extensions(self, value):
        self._file_extensions = value
