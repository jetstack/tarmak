.. _design_cli_ux:

******************************
Using Tarmak command-line tool
******************************

$ tarmak

kubectl
#######

`kubectl` on clusters (Alias for ``$ tarmak clusters kubectl``)

Usage:
  $ tarmak kubectl

------------

init
####

Inits a provider if not existing, inits an env if not existing, inits a cluster

Usage:
  $ tarmak init

-------------

Resources
#########

Tarmak has 3 resources that can be acted upon - environments, providers and clusters.

Usage:
  $ tarmak [providers | environments | clusters] [command]

-------------

providers
#########

Providers resource subcommand

list
*****************

List providers resource

Usage:
  $ tarmak providers list

init
*****************

Init providers resource

Usage:
  $ tarmak providers init

------------

environments
############

Environments resource subcommand

list
*****************

List environments resource

Usage:
  $ tarmak environments list

init
*****************

Init environments resource

Usage:
  $ tarmak environments init

------------

clusters
########

Clusters resource subcommand

list
*****************

List clusters resource

Usage:
  $ tarmak clusters list

init
*****************

Init cluster resource

Usage:
  $ tarmak clusters init

kubectl
*****************

`kubectl` on clusters resource

Usage:
  $ tarmak clusters kubectl

ssh <instance_name>
*******************

Secure Shell into an instance on clusters

Usage:
  $ tarmak clusters ssh <instance_name>

apply
*****************

Apply changes to cluster (apply infrastructure changes only)

Usage:
  $ tarmak clusters apply

plan
*****************

Dry run apply

Usage:
  $ tarmak clusters plan

XX
*****************

Does not run any infrastructure changes. Reconfigure based on configuration changes.

Usage:
  $ tarmak clusters XX

YY
*****************

Reconfigure based on infrastructure+configuration changes.

Usage:
  $ tarmak clusters YY

YY-rolling-update
*****************

YY with rolling update

Usage:
  $ tarmak clusters YY-rolling-update

instances [ list | ssh ]
************************

Instances on Cluster resource

- list

 Lists nodes of the context.

- ssh

 Alias for ``$ tarmak clusters ssh``

Usage:
  $ tarmak clusters instances [list | ssh]

server-pools [ list ]
*********************
- list

 List server pools on Cluster resource

Usage:
  $ tarmak clusters server-pools list

images [ list | build ]
***********************
- list

 List images on Cluster resource

- build

 Build images of Cluster resource

Usage:
  $ tarmak clusters images [list | build]

debug [ terraform shell | puppet | etcd | vault ]
*************************************************
- terraform shell

- puppet

- etcd

- vault

Usage:
  $ tarmak clusters debug [terraform shell | puppet | etcd | vault]

------------

Relationships
#############

The relationship between Providers, Environments and Clusters is as follows:

Provider (many) -> Environment (one)

Environment (many) -> Cluster (one)

Changed Names
#############

+-----------+-------------+
| Old Name  | New Name    |
+===========+=============+
| NodeGroup | Server Pool |
+-----------+-------------+
| Context   | Cluster     |
+-----------+-------------+
|  Nodes    | Instances   |
+-----------+-------------+
