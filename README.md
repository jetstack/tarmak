![tarmak logo](/docs/static/logo-tarmak-400px.png)

## What is Tarmak?

Tarmak is an open-source toolkit for Kubernetes cluster lifecycle management
that focuses on best practice cluster security and cluster
management/operation. It has been built from the ground-up to be cloud
provider-agnostic and hence provides a means for consistent and reliable
cluster deployment and management, across clouds and on-premises environments.

Tarmak and its underlying components are the product of
[Jetstack](https://www.jetstack.io)'s work with  its customers to build and
deploy Kubernetes in production at scale.

Under-the-hood, Tarmak uses a number of well-known and proven components,
including Terraform, Puppet and systemd.

## Quickstart

Get a ready built version of tarmak from [the releases
page](https://github.com/jetstack/tarmak/releases):

    $ wget https://github.com/jetstack/tarmak/releases/download/0.3.0-rc5/tarmak_0.3.0-rc5_linux_amd64
    $ mv tarmak_0.3.0-rc5_linux_amd64 tarmak
    $ chmod +x tarmak

If you want compile from source, follow the [build
guide](https://docs.tarmak.io/developer-guide.html#building-tarmak)
instead.

Now follow the [user guide](https://docs.tarmak.io/user-guide.html).

## Documentation

Full documentation, including design/architecture overview, user/developer
guides and more, is maintained at https://docs.tarmak.io/.

----

**Disclaimer** - please note that current releases of Tarmak are *alpha*
(unless explicitly marked).  Although we do not anticipate breaking changes, at
this stage this cannot be absolutely guaranteed.
