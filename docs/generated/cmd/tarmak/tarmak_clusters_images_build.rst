.. _tarmak_clusters_images_build:

tarmak clusters images build
----------------------------

build specific or all images missing

Synopsis
~~~~~~~~


build specific or all images missing

::

  tarmak clusters images build [base names] [flags]

Options
~~~~~~~

::

  -A, --all    build all images regardless whether they already exist
  -h, --help   help for build

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

* `tarmak clusters images <tarmak_clusters_images.html>`_ 	 - Operations on images

