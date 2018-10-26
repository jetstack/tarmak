.. _tarmak_clusters_images_destroy:

tarmak clusters images destroy
------------------------------

destroy remote tarmak images

Synopsis
~~~~~~~~


destroy remote tarmak images

::

  tarmak clusters images destroy [image ids] [flags]

Options
~~~~~~~

::

  -A, --all    destroy all tarmak images for this cluster
  -h, --help   help for destroy

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

