
File Watcher
============

File Watcher uses https://github.com/go-fsnotify/fsnotify to watch the
filesystem and execute a command when a file changes.

Install
-------

.. code::

    go get github.com/dnephin/filewatcher


Usage
-----

See ``filewatcher --help``


**Excludes**

File globbing patterns are used to match files. See
https://golang.org/pkg/path/filepath/#Match. Exclude paths must start with
a ``*/`` to match files in any directory.

**Commands**

Commands may include variables in the form ``${variable}`` which will be
replaced with a value based on the filename that was modified. Supported
variables are:

* ``filepath`` - the relative path to the file that changed
* ``dir`` - the directory of the file that changed
