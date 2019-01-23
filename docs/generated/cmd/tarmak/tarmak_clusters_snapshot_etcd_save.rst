.. _tarmak_clusters_snapshot_etcd_save:

tarmak clusters snapshot etcd save
----------------------------------

save etcd snapshot to target path prefix, i.e 'backup-'

Synopsis
~~~~~~~~


save etcd snapshot to target path prefix, i.e 'backup-'

::

  tarmak clusters snapshot etcd save [target path prefix] [flags]

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

* `tarmak clusters snapshot etcd <tarmak_clusters_snapshot_etcd.html>`_ 	 - Manage snapshots on remote etcd clusters

