.. _design_cli_ux:

Using Tarmak command-line tool
==============================

Tarmak has 3 resources that can be acted upon - environments, providers and clusters. Providers and environments are able to have `init` and `list` applied:

$ tarmak [environments|providers] [init|list]

Clusters have a larger set of commands that can be applied:

-----------

========
clusters
========

Clusters resource subcommand

------------

list
####

List clusters resource

------------

init
####

Init cluster resource

------------

kubectl
#######

`kubectl` on clusters resource

------------

ssh <instance_name>
###################

Secure Shell into an instance on clusters

------------

apply
#####

Apply changes to cluster (apply infrastructure changes only)

------------

plan
#####

Dry run apply

------------

XX
##

Does not run any infrastructure changes. Reconfigure based on configuration changes.

------------

YY
##

Reconfigure based on infrastructure+configuration changes.

------------

YY-rolling-update
#################

YY with rolling update

------------

instances [ list | ssh ]
########################

Instances on Cluster resource

- list

 Lists nodes of the context.

- ssh

 Alias for ``$ tarmak clusters ssh``

------------

server-pools [ list ]
#####################
- list

  List server pools on Cluster resource

------------

images [ list | build ]
#######################
- list

 List images on Cluster resource

- build

 Build images of Cluster resource

------------

debug [ terraform shell | puppet | etcd | vault ]
#################################################
- terraform shell

- puppet

- etcd

- vault

Relationships
#############

The relationship between Providers, Environments and Clusters is as follows:

Provider (many) -> Environment (one)

Environment (many) -> Cluster (one)

Changed Names
#############

+-----------+-------------+
| Old Names | New Names   |
+===========+=============+
| NodeGroup | Server Pool |
+-----------+-------------+
| Context   | Cluster     |
+-----------+-------------+
|  Nodes    | Instances   |
+-----------+-------------+
