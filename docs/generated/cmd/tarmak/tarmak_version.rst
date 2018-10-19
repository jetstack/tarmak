.. _tarmak_version:

tarmak version
--------------

Print the version number of tarmak

Synopsis
~~~~~~~~


Print the version number of tarmak

::

  tarmak version [flags]

Options
~~~~~~~

::

  -h, --help   help for version

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

* `tarmak <tarmak.rst>`_ 	 - Tarmak is a toolkit for provisioning and managing Kubernetes clusters.

