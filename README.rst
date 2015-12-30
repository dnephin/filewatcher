
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

File globbing patterns are used to match files. See
https://golang.org/pkg/path/filepath/#Match. Exclude paths must start with
a ``*/`` to match files in any directory. 
