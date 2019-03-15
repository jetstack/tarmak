.. existing-vpc:

.. spelling::

   aws
   vpc
   subnets

Deploying into an existing AWS VPC
==================================

Tarmak has experimental support for deploying clusters into an existing AWS
VPC.

To enable this, you will need to note down the IDs for the VPC and subnets you
want to deploy to.

For example, if we have the following infrastructure (notation in terraform):

.. literalinclude:: existing-vpc/vpc.tf

Run ``tarmak init`` as normal. Before running the ``apply`` stage, add the
following annotations to your clusters network configuration (located in
``~/.tarmak/tarmak.yaml``)::

		network:
		  cidr: 10.99.0.0/16
		  metadata:
			creationTimestamp: null
			annotations:
			  tarmak.io/existing-vpc-id: vpc-xxxxxxxx
			  tarmak.io/existing-public-subnet-ids: subnet-xxxxxxxx,subnet-xxxxxxxx,subnet-xxxxxxxx
			  tarmak.io/existing-private-subnet-ids: subnet-xxxxxxxx,subnet-xxxxxxxx,subnet-xxxxxxxx

**Note** you need to add these annotations to all clusters for that VPC, this
includes hub clusters.

Now you can run ``tarmak cluster apply`` and continue as normal.

.. warning::
  Deploying Tarmak into an existing VPC will not automatically create VPC
  endpoints for AWS services. It is strongly recommended that at least an S3 VPC
  endpoint is present for your deployed cluster.
