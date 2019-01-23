.. _tarmak_clusters_debug:

tarmak clusters debug
---------------------

Operations for debugging a cluster

Synopsis
~~~~~~~~


Operations for debugging a cluster

Options
~~~~~~~

::

  -h, --help   help for debug

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

* `tarmak clusters <tarmak_clusters.html>`_ 	 - Operations on clusters
* `tarmak clusters debug puppet <tarmak_clusters_debug_puppet.html>`_ 	 - Operations for debugging Puppet configuration
* `tarmak clusters debug terraform <tarmak_clusters_debug_terraform.html>`_ 	 - Operations for debugging Terraform configuration

