.. _tarmak_clusters_snapshot_consul_save:

tarmak clusters snapshot consul save
------------------------------------

save consul cluster snapshot to target path

Synopsis
~~~~~~~~


save consul cluster snapshot to target path

::

  tarmak clusters snapshot consul save [target path] [flags]

Options
~~~~~~~

::

  -h, --help   help for save

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  -c, --config-directory string                          config directory for tarmak's configuration (default "~/.tarmak")
      --current-cluster string                           override the current cluster set in the config
      --ignore-missing-public-key-tags ssh_known_hosts   ignore missing public key tags on instances, by falling back to populating ssh_known_hosts with the first connection (default true)
      --keep-containers                                  do not clean-up terraform/packer containers after running them
      --public-api-endpoint                              Override kubeconfig to point to cluster's public API endpoint
  -v, --verbose                                          enable verbose logging
      --wing-dev-mode                                    use a bundled wing version rather than a tagged release from GitHub

SEE ALSO
~~~~~~~~

* `tarmak clusters snapshot consul <tarmak_clusters_snapshot_consul.html>`_ 	 - Manage snapshots on remote consul clusters

