.. vim:set ft=rst spell:

Managing SSH Known Hosts
========================

This proposal suggests a method to more securely manage and ensure
authentication of remote hosts through the SSH protocol within Tarmak
environments on AWS.

Background
----------

Currently we solely use the external OpenSHH SSH command line tool to connect to
remote instances on EC2 for both interactive shell sessions as well as
tunnelling and proxy commands for other services, including, connecting to the
Vault cluster and the private Kubernetes API server endpoint. Currently in
development is the replacement of our programmatic use cases of SSH in favour of
the in package Go solution, a choice stemming from pain points developing more
sophisticated utility functions for Tarmak and the desire for improvements in
control of connections to remote hosts.

During development of this replacement it became clear that proper care must be
taken during authentication of host public keys during connection and manual
management of our ``ssh_known_hosts`` cluster file. Our current implementation
allows OpenSSH to maintain this file however, does not exit with an error if
public keys do not match due to the flag ``StrictHostKeyChecking`` set to
``no``. Not only does a miss-match in public keys not cause an error, the
population of known public keys on different authenticated machines to the same
EC2 hosts will always use the hosts presented public key, meaning the set of
public keys could potentially be different for users accessing the same cluster.

Objective
---------

By implementing stricter enforcement of the ``ssh_known_hosts`` file and passing
it's management to Tarmak, we can improve the security of SSH connections to
remote hosts. The key high level points to achieving this is as follows:

 - Disable writes from the OpenSSH command to the ``ssh_known_hosts`` file and
   enforce strict checking.
 - Enforce that our in package implementation of SSH connections adheres to this
   file also.
 - Collect public keys during instance start up that are then stored, tightly
   coupled with that host. These keys are able to be used as a source of truth
   for other authenticated users attempting to connect to remote hosts on the
   cluster that have empty or an incomplete ``ssh_known_hosts`` file.

Changes
-------

Firstly, we must restrict the OpenSSH command line tool from editing the
``ssh_known_hosts`` file and strictly enforce it by updating the generator for the
``ssh_config`` file. This enables Tarmak to take control of the ``ssh_known_hosts``
file management.

In order to create a source of truth for each host's public key, each instance
will have it's public key's attached as a tags, shortly after boot time like the
following:

+-------------------------+---------------------------------------------------------------------------+
| PublicKey_ssh-ed25519_0 | AAAAC3NzaC1lZDI1NTE5AAAAIE90XYYm6GSDlNGejM+aY5dZEe5vK4XyU++89WdGJcDc==EOF |
+-------------------------+---------------------------------------------------------------------------+

The population of these tags will happen at boot time for all instances,
regardless of whether they have been created from a direct Terraform apply or
via an Amazon Auto Scaling Group. At execution time, Wing - present on every
instance - will invoke an Amazon Lambda function for Instance Tagging. Passed to
this function will be a collection of the instances public keys, it's Amazon
identity document and matching PKCS7 document.

Upon receiving this request, the lambda function will verify the authenticity
of the request and identity document by verifying the instance identity and
PKCS7 document against the public AWS certificate. Further details on this can
be found `here
<https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instance-identity-documents.html>`_.
Once verified, the function will split the public keys to maximum sized chunks
of 256 - the maximum size length of EC2 tags.

Finally, the function will test for the existence of these tags and do one of
three actions:

 - if tags exist and match, exit success
 - if tags exist and miss-match, exit failure
 - if tags do not exist, create tags and exit success

Once an instance has requested for the creation of it's tags, all subsequent
requests should succeed with no action.

All SSH connections will rely on the contents of the ``ssh_known_hosts``
file however, in the case the host is not present in the file, will attempt to
use the AWS instance's ``publicKey..`` tag to populate it's entry.

Notable items
-------------

A start has been made on the code for the Lambda function:

.. code-block:: go

  package main

  import (
  	"context"
  	"fmt"
  	"time"

  	"github.com/aws/aws-lambda-go/lambda"
  )

  const (
  	AWSCert = "global aws public cert"
  	tagSize = 256
  )

  type EC2InstanceIdentityDocument struct {
  	DevpayProductCodes []string  `json:"devpayProductCodes"`
  	AvailabilityZone   string    `json:"availabilityZone"`
  	PrivateIP          string    `json:"privateIp"`
  	Version            string    `json:"version"`
  	Region             string    `json:"region"`
  	InstanceID         string    `json:"instanceId"`
  	BillingProducts    []string  `json:"billingProducts"`
  	InstanceType       string    `json:"instanceType"`
  	AccountID          string    `json:"accountId"`
  	PendingTime        time.Time `json:"pendingTime"`
  	ImageID            string    `json:"imageId"`
  	KernelID           string    `json:"kernelId"`
  	RamdiskID          string    `json:"ramdiskId"`
  	Architecture       string    `json:"architecture"`
  }

  type TagInstanceRequest struct {
  	PublicKeys       map[string][]byte           `json:"publicKeys"`
  	InstanceDocument EC2InstanceIdentityDocument `json:"instanceID"`
  	PKCS7CMS         string                      `json:"pkcs7CMS"`
  }

  func HandleRequest(ctx context.Context, t TagInstanceRequest) error {
  	if err := t.verify(); err != nil {
  		return err
  	}

  	tags := t.createTags()

  	exists, err := t.checkTagsAgainstInstance(tags)
  	if err != nil || exists {
  		return err
  	}

  	// attach tags to ec2 instance using real call
  	//err := ec2.Tag{
  	//	InstanceID: t.InstanceDocument.InstanceID,
  	//	Tags: ....
  	//}
  	// if err != nil {
  	//	return err
  	//}

  	return nil
  }

  // verify the pkcs7 doc against the instance identity content and AWS global
  // cert
  func (t TagInstanceRequest) verify() error {
  	return nil
  }

  // check generated tags against the ec2 instance
  // if existing and match exit gracefully
  // if miss match, exit failure
  // if not exist, we need to create
  func (t TagInstanceRequest) checkTagsAgainstInstance(tags map[string][]byte) (tagsExist bool, err error) {
  	return false, nil
  }

  // split up public keys into correct sizes for AWS tags
  func (t TagInstanceRequest) createTags() map[string][]byte {
  	tags := make(map[string][]byte)

  	for keyName, data := range t.PublicKeys {
  		data = append(data, []byte("==EOF")...)

  		for i := 0; i < len(data); i += tagSize {
  			end := i + tagSize

  			if end > len(data) {
  				end = len(data)
  			}

  			tagName := fmt.Sprintf("PublicKey_%s_%s", keyName, i/tagSize)
  			tags[tagName] = data[i:end]
  		}
  	}

  	return tags
  }

  func main() {
  	lambda.Start(HandleRequest)
  }

Out of scope
------------

We should not disrupt the current flow of key generation on the host instances
such as using key injection. At no point should private keys be in flight.

We should not store or rely on the public key being stored in the Terraform
state as this would require all commands that rely on SSH, to also rely on
fetching and updating the Terraform state - significantly increasing completion
time for even trivial tasks.
