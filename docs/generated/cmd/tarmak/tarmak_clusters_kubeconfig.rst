.. _tarmak_clusters_kubeconfig:

tarmak clusters kubeconfig
--------------------------

Verify and get path to Kubeconfig

Synopsis
~~~~~~~~


Verify and get path to Kubeconfig

::

  tarmak clusters kubeconfig [flags]

Options
~~~~~~~

::

  -h, --help                  help for kubeconfig
  -p, --path string           Path to save your kubeconfig file to
      --public-api-endpoint   Use public API endpoint for your kubeconfig

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

* `tarmak clusters <tarmak_clusters.html>`_ 	 - Operations on clusters

