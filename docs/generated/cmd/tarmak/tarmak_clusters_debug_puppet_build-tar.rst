.. _tarmak_clusters_debug_puppet_build-tar:

tarmak clusters debug puppet build-tar
--------------------------------------

Build a puppet.tar.gz in the current working directory

Synopsis
~~~~~~~~


Build a puppet.tar.gz in the current working directory

::

  tarmak clusters debug puppet build-tar [flags]

Options
~~~~~~~

::

  -h, --help   help for build-tar

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  -c, --config-directory string        config directory for tarmak's configuration (default "~/.tarmak")
      --current-cluster string         override the current cluster set in the config
      --keep-containers                do not clean-up terraform/packer containers after running them
      --log-flush-frequency duration   Maximum number of seconds between log flushes (default 5s)
  -v, --verbose                        enable verbose logging
      --wing-dev-mode                  use a bundled wing version rather than a tagged release from GitHub

SEE ALSO
~~~~~~~~

* `tarmak clusters debug puppet <tarmak_clusters_debug_puppet.rst>`_ 	 - Operations for debugging Puppet configuration

