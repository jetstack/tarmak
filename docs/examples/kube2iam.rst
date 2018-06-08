Kube2IAM
--------

Setup
~~~~~

Create instance IAM policy
++++++++++++++++++++++++++

Every instance that will run the kube2iam pod needs to have an specific
IAM policy attached to the IAM role of that instance.

The following little Terraform project creates an IAM policy that will
give instances the ability to assume roles. We limit the access to which
roles the instances has access, by only allowing it to access roles in a 
restricted path in AWS IAM.
The Terraform project has 2 inputs ``aws_region`` and ``cluster_name``.
It also has 2 outputs defined the ``ARN`` and ``path`` of the policy.
The ARN is what you need to give to Tarmak and the path is needed to be
able to deploy your roles for the pods in the correct path.

.. code-block:: none

    terraform {}

    provider "aws" {
    region = "${var.aws_region}"
    }

    variable "aws_region" {
    description = "AWS Region you want to deploy it in"
    default     = "eu-west-1"
    }

    variable "cluster_name" {
    description = "Name of the cluster"
    }

    data "aws_caller_identity" "current" {}

    resource "aws_iam_policy" "kube2iam" {
    name        = "kube2iam_assumeRole_policy_${var.cluster_name}"
    path        = "/"
    description = "Kube2IAM role policy for ${var.cluster_name}"

    policy = "${data.aws_iam_policy_document.kube2iam.json}"
    }

    data "aws_iam_policy_document" "kube2iam" {
    statement {
        sid = "1"

        actions = [
        "sts:AssumeRole",
        ]

        effect = "Allow"

        resources = [
        "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/kube2iam_${var.cluster_name}/*",
        ]
    }
    }

    output "kube2iam_arn" {
    value = "${aws_iam_policy.kube2iam.arn}"
    }

    output "kube2iam_path" {
    value = "${aws_iam_policy.kube2iam.path}"
    }


You can run the Terraform project the following way:

.. code-block:: bash

    terraform init
    terraform apply -var cluster_name=example -var region=eu-west-1

Attach instance policy
++++++++++++++++++++++

Add the created IAM policy ARN to your tarmak config. You can do this by
adding additional IAM policies.

Deploy kube2iam
+++++++++++++++

With `HELM <https://www.helm.sh/>`_ it is really easy to deploy kube2iam 
with the correct settings.

You can deploy it with the following command:

.. code-block:: bash

    helm upgrade kube2iam stable/kube2iam \
    --install \
    --version 0.10.0 \
    --namespace kube-system \
    --set=extraArgs.host-ip=127.0.0.1 \
    --set=extraArgs.log-format=json \
    --set=updateStrategy=RollingUpdate \
    --set=rbac.create=true \
    --set=host.iptables=false


We set ``iptables`` to false and ``host-ip`` to 127.0.0.1 as Tarmak already created
the iptables rule and forward it to ``127.0.0.1:8181``.
Specific kube2iam options can be found in the `documentation <https://github.com/jtblin/kube2iam>`_ of kube2iam.

Usage
~~~~~

Now that kube2IAM is installed on your system, you can start creating roles
and policies to give your pods access to AWS resources.

An example creation of an IAM policy and role:

.. code-block:: none

    resource "aws_iam_role" "test_role" {
    name = "test_role"
    path = "/kube2iam_example/"

    assume_role_policy = <<EOF
    {
    "Version": "2012-10-17",
    "Statement": [
        {
        "Action": "sts:AssumeRole",
        "Principal": {
        "AWS": [
            "<arn of instance profile role>"
        ]
        },
        "Effect": "Allow"
        }
    ]
    }
    EOF
    }

    resource "aws_iam_role_policy" "test_role_policy" {
    name = "test_policy"
    role = "${aws_iam_role.test_role.id}"

    policy = <<EOF
    {
    "Version": "2012-10-17",
    "Statement": [
        {
        "Effect": "Allow",
        "Action": [
            "s3:ListBucket"
        ],
        "Resource": [
            "*"
        ]
        }
    ]
    }
    EOF
    }

    output "test_role" {
    value = "${aws_iam_role.test_role.arn}"
    }

When you create a role, you need to make sure you deploy it in the correct
``path`` and also add an assume role policy to it. That assume role policy
needs to grant access to the role ARN that is attached to the instances.
