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

For example, if we have the following infrastructure (notation in terraform)::

		provider "aws" {}

		data "aws_availability_zones" "available" {}

		resource "aws_vpc" "main" {
		  cidr_block = "10.0.0.0/16"
		  enable_dns_support   = true
		  enable_dns_hostnames = true

		  tags {
			Name = "test_vpc"
		  }
		}

		resource "aws_eip" "nat" {
		  vpc = true
		}

		resource "aws_subnet" "public" {
		  count             = "${length(data.aws_availability_zones.available.names)}"
		  vpc_id            = "${aws_vpc.main.id}"
		  cidr_block        = "${cidrsubnet(cidrsubnet(aws_vpc.main.cidr_block, 3, 0), 3, count.index)}"
		  availability_zone = "${data.aws_availability_zones.available.names[count.index]}"

		  tags {
			Name = "public_${data.aws_availability_zones.available.names[count.index]}"
		  }
		}

		resource "aws_subnet" "private" {
		  count             = "${length(data.aws_availability_zones.available.names)}"
		  vpc_id            = "${aws_vpc.main.id}"
		  cidr_block        = "${cidrsubnet(aws_vpc.main.cidr_block, 3, count.index + 1)}"
		  availability_zone = "${data.aws_availability_zones.available.names[count.index]}"

		  tags {
			Name = "private_${data.aws_availability_zones.available.names[count.index]}"
		  }
		}

		resource "aws_internet_gateway" "main" {
		  vpc_id = "${aws_vpc.main.id}"
		}

		resource "aws_nat_gateway" "main" {
		  count         = "${length(aws_subnet.public)}"
		  depends_on    = ["aws_internet_gateway.main"]
		  allocation_id = "${aws_eip.nat.id}"
		  subnet_id     = "${aws_subnet.public.*.id[count.index]}"
		}

		resource "aws_route_table" "public" {
		  vpc_id = "${aws_vpc.main.id}"
		}

		resource "aws_route" "public" {
		  route_table_id         = "${aws_route_table.public.id}"
		  destination_cidr_block = "0.0.0.0/0"
		  gateway_id             = "${aws_internet_gateway.main.id}"
		}

		resource "aws_route_table" "private" {
		  vpc_id = "${aws_vpc.main.id}"
		}

		resource "aws_route" "private" {
		  route_table_id         = "${aws_route_table.private.id}"
		  destination_cidr_block = "0.0.0.0/0"
		  nat_gateway_id         = "${aws_nat_gateway.main.id}"
		}

		resource "aws_route_table_association" "public" {
		  count          = "${length(data.aws_availability_zones.available.names)}"
		  subnet_id      = "${aws_subnet.public.*.id[count.index]}"
		  route_table_id = "${aws_route_table.public.id}"
		}

		resource "aws_route_table_association" "private" {
		  count          = "${length(data.aws_availability_zones.available.names)}"
		  subnet_id      = "${aws_subnet.private.*.id[count.index]}"
		  route_table_id = "${aws_route_table.private.id}"
		}

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
