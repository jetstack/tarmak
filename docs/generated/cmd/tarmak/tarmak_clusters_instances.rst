.. _tarmak_clusters_instances:

tarmak clusters instances
-------------------------

Operations on instances

Synopsis
~~~~~~~~


Operations on instances

Options
~~~~~~~

::

  -h, --help   help for instances

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
* `tarmak clusters instances list <tarmak_clusters_instances_list.html>`_ 	 - Print a list of instances in the cluster
* `tarmak clusters instances ssh <tarmak_clusters_instances_ssh.html>`_ 	 - Log into an instance with SSH

