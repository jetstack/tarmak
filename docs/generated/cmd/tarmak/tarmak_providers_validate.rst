.. _tarmak_providers_validate:

tarmak providers validate
-------------------------

Validate provider(s) used by current cluster

Synopsis
~~~~~~~~


Validate provider(s) used by current cluster

::

  tarmak providers validate [flags]

Options
~~~~~~~

::

  -h, --help   help for validate

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

* `tarmak providers <tarmak_providers.rst>`_ 	 - Operations on providers

