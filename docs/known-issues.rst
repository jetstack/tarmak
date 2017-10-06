.. known-issues:

.. spelling::

   aws
   arn
   kms

Known Issues
============

This document summarises some of the known issues users may come across when running Tarmak and how to deal with them.

An alias with the name arn:aws:kms:<region>:<id>:alias/tarmak/<environment>/secrets already exists
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

If you lose your terraform state file after spinning up a cluster with Tarmak, terraform cannot delete anything that was in that state file. On the next run of Tarmak, terraform will try to recreate the resources required for your cluster. One such resource is AWS KMS aliases, which need to be unique and cannot be deleted through the AWS console. In order to delete these aliases manually based on the error above you can run:

::

  aws kms delete-alias --region <region>  --alias-name alias/tarmak/<environment>/secrets

aws key pair is not matching the local one
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

If you run into following error when running ``tarmak clusters apply``:

:: 
  
  FATA[0004] error preparing container: error validating environment: 1 error occurred:

  * aws key pair is not matching the local one, aws_fingerprint=<aws_fingerprint> local_fingerprint=<local_fingerprint>

then there is a mismatch between the your key pair's public key stored by AWS and your local key pair. To remedy this you must either create a new key pair and upload it to AWS manually, or delete your existing key pair through the AWS console and re-run ``tarmak clusters apply``.

