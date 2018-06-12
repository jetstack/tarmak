.. vim:set ft=rst spell:

#######################
Puppet Dry Run Via Wing
#######################

Background
==========

Using puppet's noop mode we can perform a proper dry-run of the changes that be
performed on cluster instances.

See http://nrvale0.github.io/posts/the-basics-of-puppet-noop/ for an overview on noop mode.

Glossary
--------

- **Puppet dry-run**: a puppet mode which calculates changes, but does not actually apply them to the instance.

Problem
-------

Currently there is no way to verify what a ``puppet apply`` will do to nodes.
In some situations, we would like to review the changes that Puppet would do on
the system without actually applying them, to prevent causing destructive
configuration changes.

We need a way to represent applying a set of puppet manifests to an instance,
and a way to view the output of an application.

See:

- https://github.com/jetstack/tarmak/issues/224

Objective
=========

Verify puppet will make sensible and expected changes to the cluster when running::

    $ tarmak cluster config-plan

which complements ``cluster plan`` verifying Terraform changes.

Changes
=======

To implement this, new objects will be added to Wing's API.

New API objects
---------------

``PuppetTarget``
****************

A resource representing a set of puppet modules and additional information
(hieradata etc) to apply to an instance.

This bundles together the source (S3, google Cloud Storage, etc) and
verification of the source (a hash, PGP signature, etc).

.. code-block:: yaml

    kind: PuppetTarget
    metadata:
      name: example-manifest
    hash: sha256:34242343
    source:
      s3:
        bucketName: something
        path: something-else/puppet.tar.gz

The source will be structured similarly to kubernetes' ``VolumeSource``, with a
field for each type of source. For example, something like this in ``types.go``:

.. code-block:: go

    type ManifestSource struct {
           S3 *S3ManifestSource `json:"s3"`
    }

    type S3ManifestSource struct {
           BucketName string `json:"bucketName"`
    }


``WingJob``
*************

A resource representing the application of a ``PuppetTarget`` on an instance:

.. code-block:: yaml

    kind: WingJob
    metadata:
      name: example-job
    spec:
      puppetTargetRef:
        name: example-manifest
      operation: "dry-run"
      instanceID: 1234
    status:
      exitCode: 1
      messages: ""

This references a pre-existing ``PuppetTarget``, and performs the specified
action on an instance.

If, in the future, we support other configuration management tools (ansible,
chef, etc), these would be represented by separate fields.

Performing updates to puppet manifests will leave ``WingJob`` and
``PuppetTarget`` resources hanging around. To prevent this, wing should only
keep the last 15 (or some other number) WingJobs for each instance.

Changes to existing API objects
-------------------------------

``InstanceSpec`` will have a ``spec.puppetTargetRef`` field also linking to a
``PuppetTarget`` resource.  This will be the manifest applied to the instance.
A corresponding ``status.puppetTargetRef`` will also be added to record the
latest known applied ``PuppetTarget``.

When the ``spec.puppetTargetRef`` field is updated on an instance, a new
``WingJob`` is created with ``operation`` set to ``apply``.

If a ``WingJob`` applies a different ``PuppetTarget``, the
``spec.puppetTargetRef`` will not be updated, however
``status.puppetTargetRef`` will be.

Changes to tarmak CLI
---------------------

The tarmak CLI needs modification to add support for creating
``PuppetTarget`` and ``PuppetJob`` resources.

The planned workflow is to run::

    $ tarmak cluster config-plan

which creates ``WingJob`` resources for either a subset of instances of each
type in the current cluster, or all instances. This blocks until
``WingJob.Status.ExitCode`` for each created job is populated.

It would be nice to filter and only display results based on the exit code of puppet, but it seems the exit code is always ``0`` when ``--noop`` is enabled::

    https://tickets.puppetlabs.com/browse/PUP-686

Notable items
=============

Concerns
--------

- We need to think about how to handle ``WingJob`` resources timing out in the case of an instance failure during a plan.
