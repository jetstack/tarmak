AWS key pair is not matching the local one
------------------------------------------

If you run into following error when running ``tarmak clusters apply``:

::

  FATA[0004] error preparing container: error validating environment: 1 error occurred:

  * aws key pair is not matching the local one, aws_fingerprint=<aws_fingerprint> local_fingerprint=<local_fingerprint>

then there is a mismatch between your key pair's public key stored by AWS and
your local key pair. To remedy this you must either create a new key pair and
upload it to AWS manually, or delete your existing key pair through the AWS
console and re-run ``tarmak clusters apply``.
