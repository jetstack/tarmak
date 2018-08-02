.. _design_cli_ux:

***************************
Command-line tool reference
***************************

Here are the commands and resources for the ``tarmak`` command-line tool.

Commands
--------

``kubectl``
~~~~~~~~~~~

Run ``kubectl`` on clusters (Alias for ``$ tarmak clusters kubectl``).

Usage::

  $ tarmak kubectl

------------

``init``
~~~~~~~~

* Initialises a provider if not existing.
* Initialises an environment if not existing.
* Initialises a cluster.

Usage::

  $ tarmak init

-------------

Resources
---------

Tarmak has three resources that can be acted upon - environments, providers and clusters.

Usage::

  $ tarmak [providers | environments | clusters] [command]

-------------

Providers
~~~~~~~~~

Providers resource sub-command.

``list``
********

List providers resource.

Usage::

  $ tarmak providers list

``init``
********

Initialise providers resource.

Usage::

  $ tarmak providers init

------------

Environments
~~~~~~~~~~~~

Environments resource sub-command.

``list``
********

List environments resource.

Usage::

  $ tarmak environments list

``init``
********

Initialise environments resource.

Usage::

  $ tarmak environments init

------------

Clusters
~~~~~~~~

Clusters resource sub-command.

``list``
********

List clusters resource.

Usage::

  $ tarmak clusters list

``init``
********

Initialise cluster resource.

Usage::

  $ tarmak clusters init



``set-context``
***************

Change active cluster.

Usage::

  $ tarmak clusters set-context [EnvironmentName-ClusterName]


``kubectl``
***********

Run ``kubectl`` on clusters resource.

Usage::

  $ tarmak clusters kubectl

``ssh <instance_name>``
***********************

Secure Shell into an instance on clusters.

Usage::

  $ tarmak clusters ssh <instance_name>

``apply``
*********

Apply changes to a cluster (by default applies infrastructure (Terraform) and configuration (Puppet) changes.

Usage::

  $ tarmak clusters apply

Flags::

  --infrastructure-stacks [state,network,tools,vault,kubernetes]
      target exactlyone piece of the infrastructure (aka terraform stack). This implies (--infrastructure-only)
  --infrastructure-only   [default=false]
      only apply infrastructure (aka terraform)
  --configuration-only    [default=false]
      only apply configuration  (aka puppet)
  --dry-run               [default=false]
      show changes only, do not actually execute them

``destroy``
***********

Destroy the infrastructure of a cluster

Usage::

  $ tarmak clusters destroy

Flags::

  --infrastructure-stacks     [state,network,tools,vault,kubernetes]
      target exactlyone piece of the infrastructure (aka terraform stack). This implies (--infrastructure-only)
  --force-destroy-state-stack [default=false]
      force destroy the state stack, this is unreversible
  --dry-run                   [default=false]
      show changes only, do not actually execute them


``instances [ list | ssh ]``
****************************

Instances on Cluster resource.

``list``
^^^^^^^^

Lists nodes of the context.

``ssh``
^^^^^^^

Alias for ``$ tarmak clusters ssh``.

Usage::

  $ tarmak clusters instances [list | ssh]

``server-pools [ list ]``
*************************

``list``
^^^^^^^^

List server pools on Cluster resource.

Usage::

  $ tarmak clusters server-pools list

``images [ list | build ]``
***************************

``list``
^^^^^^^^

List images on Cluster resource.

``build``
^^^^^^^^^

Build images of Cluster resource.

Usage::

  $ tarmak clusters images [list | build]

``debug [ terraform shell | puppet | etcd | vault ]``
*****************************************************

Used for debugging.

``terraform shell``
^^^^^^^^^^^^^^^^^^^

Debug terraform via shell.

Usage::

  $ tarmak clusters debug terraform [shell]

``puppet``
^^^^^^^^^^

Debug puppet.

Usage::

  $ tarmak clusters debug puppet []

``etcd``
^^^^^^^^

Debug etcd.

Usage::

  $ tarmak clusters debug etcd [status|shell|etcdctl]

``vault``
^^^^^^^^^

Debug vault.

Usage::

  $ tarmak clusters debug vault [status|shell|vault]
