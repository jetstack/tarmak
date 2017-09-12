.. _design_cli_ux:

Using Tarmak command-line tool
==============================

Tarmak has 3 resources that can be acted upon - environments, providers and clusters. Providers and environments are able to have `init` and `list` applied:

$ tarmak [environments|providers] [init|list]

Clusters have a larger set of commands that can be applied:
  clusters
    - `list`
    - `init`
    - `kubectl`
    - `ssh <instance_name>`
    - `apply` (apply infrastructure changes only)
    - `plan`  (dry run apply)
    - `XX` (not run any infrastructure changes, reconfigure based on configuration changes)
    - `YY` (reconfigure based on infrastructure+configuration changes)
    - `YY-rolling-update`

    - instances

      - `list` (list nodes of the context)

      - `ssh` (same as ``$ tarmak clusters ssh``)
    - server-pools

      - `list`

    - images

      - `list`

      - `build`

    - debug

      - terraform

        - `shell`

      - puppet
      - etcd
      - vault

Relationships
-------------

The relationship between Providers, Environments and Clusters is as follows:

Provider (many) -> Environment (one)

Environment (many) -> Cluster (one)

Changed Names
-------------

+-----------+-------------+
| Old Names | New Names   |
+===========+=============+
| NodeGroup | Server Pool |
+-----------+-------------+
| Context   | Cluster     |
+-----------+-------------+
|  Nodes    | Instances   |
+-----------+-------------+
