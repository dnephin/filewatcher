
File Watcher
============

File Watcher uses https://github.com/go-fsnotify/fsnotify to watch the
filesystem and execute a command when a file changes.

Install
-------

.. code::

    go get github.com/dnephin/filewatcher

Examples
--------

**Run go tests**

Run go tests for a package when a file is modified, exclude vim swap files.

.. code::

    filewatcher -x '**/*.swp' -x vendor/ -x .git go test './${dir}'


Usage
-----

See ``filewatcher --help``


**Excludes**

File globbing patterns are used to match files, with one addition.
See https://golang.org/pkg/path/filepath/#Match for the standard file matching
rules.  Exclude paths may also use a ``**/`` prefix, which matches the pattern
against any directory. This may be used to ignore files with a specific
extension that may occur in any directory in the hierarhcy.

**Commands**

Commands may include variables in the form ``${variable}`` which will be
replaced with a value based on the file that was modified. Supported
variables are:

* ``filepath`` - the relative path to the file that changed
* ``dir`` - the directory of the file that changed
* ``relative_dir`` - the directory of the file that changes with a ./ prefix

These values are also set as environment variables for the process:

* ``TEST_FILENAME`` - the relative path to the file that changed
* ``TEST_DIRECTORY`` - the relative path to the directory of the file that
  changed
