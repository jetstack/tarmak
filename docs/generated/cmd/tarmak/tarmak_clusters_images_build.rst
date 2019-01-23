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

  -h, --help               help for build
  -R, --rebuild-existing   build all images regardless whether they already exist

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

* `tarmak clusters images <tarmak_clusters_images.html>`_ 	 - Operations on images

