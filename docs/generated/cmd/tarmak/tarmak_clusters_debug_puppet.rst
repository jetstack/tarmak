.. _tarmak_clusters_debug_puppet:

tarmak clusters debug puppet
----------------------------

Operations for debugging Puppet configuration

Synopsis
~~~~~~~~


Operations for debugging Puppet configuration

Options
~~~~~~~

::

  -h, --help   help for puppet

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

* `tarmak clusters debug <tarmak_clusters_debug.html>`_ 	 - Operations for debugging a cluster
* `tarmak clusters debug puppet build-tar <tarmak_clusters_debug_puppet_build-tar.html>`_ 	 - Build a puppet.tar.gz in the current working directory

