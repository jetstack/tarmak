An alias with the name ``arn:aws:kms:<region>:<id>:alias/tarmak/<environment>/secrets`` already exists
------------------------------------------------------------------------------------------------------

If you lose your terraform state file after spinning up a cluster with Tarmak, terraform cannot delete anything that was in that state file. On the next run of Tarmak, terraform will try to recreate the resources required for your cluster. One such resource is AWS KMS aliases, which need to be unique and cannot be deleted through the AWS console. In order to delete these aliases manually based on the error above you can run:

::

  aws kms delete-alias --region <region>  --alias-name alias/tarmak/<environment>/secrets


