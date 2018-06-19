Kube2IAM
--------

Kube2IAM is an extension to kubernetes that will allow you to give
fine grained AWS access to pods without. More information about the
project can be found on the `project page <https://github.com/jtblin/kube2iam>`_.

Prerequisite
~~~~~~~~~~~~

Make sure `HELM <https://www.helm.sh/>`_ is `activated <https://docs.tarmak.io/user-guide.html#tiller>`_ on the Tarmak cluster.
You also need to make sure you can connect to the cluster with your HELM install.

.. code-block:: bash

    helm version
    Client: &version.Version{SemVer:"v2.9.1", GitCommit:"20adb27c7c5868466912eebdf6664e7390ebe710", GitTreeState:"clean"}
    Server: &version.Version{SemVer:"v2.9.1", GitCommit:"20adb27c7c5868466912eebdf6664e7390ebe710", GitTreeState:"clean"}

Setup
~~~~~

Create instance IAM policy
++++++++++++++++++++++++++

Every instance that will run the kube2iam pod needs to have an specific
IAM policy attached to the IAM instance role of that instance.

The following Terraform project creates an IAM policy that will give 
instances the ability to assume roles. The assume role is restricted to
only have access to roles deployed in a specific path. By doing this, we can
limit the amount of roles an instance can assume to only the roles that it really
needs to.
The Terraform project has 2 inputs ``aws_region`` and ``cluster_name``.
The projects also has 2 outputs defined the ``ARN`` and ``path`` of the IAM policy.
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
adding `additional IAM policies <https://docs.tarmak.io/user-guide.html#additional-iam-policies>`_.

Deploy kube2iam
+++++++++++++++

With `HELM <https://www.helm.sh/>`_ it is easy to deploy kube2iam with 
the correct settings.

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


We set ``iptables`` to false and ``host-ip`` to 127.0.0.1 as Tarmak already creates
the iptables rule and forward it to ``127.0.0.1:8181``.
Specific kube2iam options can be found in the `documentation <https://github.com/jtblin/kube2iam>`_ of kube2iam.

Usage
~~~~~

Now that kube2IAM is installed on your system, you can start creating roles
and policies to give your pods access to AWS resources.

An example creation of an IAM policy and role:

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

    variable "instance_iam_role_arn" {
        description = "ARN of the instance IAM role
    }


    resource "aws_iam_role" "test_role" {
        name = "test_role"
        path = "/kube2iam_${var.cluster_name}/"

        assume_role_policy = <<EOF
        {
        "Version": "2012-10-17",
        "Statement": [
            {
            "Action": "sts:AssumeRole",
            "Principal": {
            "AWS": [
                "${instance_iam_role_arn}"
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

Now you can run this Terraform project the following way:

.. code-block:: bash

    terraform init
    terraform apply -var cluster_name=example -var region=eu-west-1 -var instance_arn=arn:aws:iam::xxxxxxx:role/my-instance-role


When you create a role, you need to make sure you deploy it in the correct
``path`` and also add an assume role policy to it. That assume role policy
needs to grant access to the role ARN that is attached to the instances.
In our example Terraform project above we solved that by adding a variable for
the ``instance_arn`` and the ``cluster_name``

With the output of the test role, you can add that as an annotation to your deployment.

.. code-block:: yaml

    apiVersion: extensions/v1beta1
    kind: Deployment
    metadata:
        name: nginx-deployment
    spec:
        replicas: 3
        template:
            metadata:
                annotations:
                    iam.amazonaws.com/role: role-arn
                labels:
                    app: nginx
            spec:
                containers:
                - name: nginx
                  image: nginx:1.9.1
                  ports:
                  - containerPort: 80
