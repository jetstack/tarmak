.. _tarmak_clusters_apply:

tarmak clusters apply
---------------------

Create or update the currently configured cluster

Synopsis
~~~~~~~~


Create or update the currently configured cluster

::

  tarmak clusters apply [flags]

Options
~~~~~~~

::

      --auto-approve                 auto approve to responses when applying cluster (default true)
      --auto-approve-deleting-data   auto approve deletion of any data as a cause from applying cluster (default true)
  -C, --configuration-only           apply changes to configuration only, by running only puppet
      --dry-run                      don't actually change anything, just show changes that would occur
  -h, --help                         help for apply
  -I, --infrastructure-only          apply changes to infrastructure only, by running only terraform
  -P, --plan-file-location string    location of stored terraform plan executable file to be used (default "${TARMAK_CONFIG}/${CURRENT_CLUSTER}/terraform/tarmak.plan")

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  -c, --config-directory string        config directory for tarmak's configuration (default "~/.tarmak")
      --current-cluster string         override the current cluster set in the config
      --keep-containers                do not clean-up terraform/packer containers after running them
      --log-flush-frequency duration   Maximum number of seconds between log flushes (default 5s)
      --public-api-endpoint            Override kubeconfig to point to cluster's public API endpoint
  -v, --verbose                        enable verbose logging
      --wing-dev-mode                  use a bundled wing version rather than a tagged release from GitHub

SEE ALSO
~~~~~~~~

* `tarmak clusters <tarmak_clusters.html>`_ 	 - Operations on clusters

