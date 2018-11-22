BUNDLE_DIR ?= .bundle

VERSION_1_13 := 1.13.0-beta.1
VERSION_1_12 := 1.12.2
VERSION_1_11 := 1.11.4
VERSION_1_10 := 1.10.10
VERSION_1_9 := 1.9.11
VERSION_1_8 := 1.8.15
VERSION_1_7 := 1.7.16
VERSION_1_6 := 1.6.13
VERSION_1_5 := 1.5.8

verify: bundle_install
	bundle exec rake test

bundle_install:
	bundle install --path $(BUNDLE_DIR)

acceptance: acceptance-1-11-centos

acceptance-1-5-centos: export KUBERNETES_VERSION = $(VERSION_1_5)
acceptance-1-5-centos: export KUBERNETES_AUTHORIZATION_MODE = ['RBAC']
acceptance-1-5-centos: bundle_install
	bundle exec rake beaker:default

acceptance-1-6-centos: export KUBERNETES_VERSION = $(VERSION_1_6)
acceptance-1-6-centos: bundle_install
	bundle exec rake beaker:default

acceptance-1-7-centos: export KUBERNETES_VERSION = $(VERSION_1_7)
acceptance-1-7-centos: bundle_install
	bundle exec rake beaker:default

acceptance-1-8-centos: export KUBERNETES_VERSION = $(VERSION_1_8)
acceptance-1-8-centos: bundle_install
	bundle exec rake beaker:default

acceptance-1-9-centos: export KUBERNETES_VERSION = $(VERSION_1_9)
acceptance-1-9-centos: bundle_install
	bundle exec rake beaker:default

acceptance-1-10-centos: export KUBERNETES_VERSION = $(VERSION_1_10)
acceptance-1-10-centos: bundle_install
	bundle exec rake beaker:default

acceptance-1-11-centos: export KUBERNETES_VERSION = $(VERSION_1_11)
acceptance-1-11-centos: bundle_install
	bundle exec rake beaker:default

acceptance-1-12-centos: export KUBERNETES_VERSION = $(VERSION_1_12)
acceptance-1-12-centos: bundle_install
	bundle exec rake beaker:default

acceptance-1-13-centos: export KUBERNETES_VERSION = $(VERSION_1_13)
acceptance-1-13-centos: bundle_install
	bundle exec rake beaker:default

acceptance-1-7-ubuntu: export KUBERNETES_VERSION = $(VERSION_1_7)
acceptance-1-7-ubuntu: bundle_install
	bundle exec rake beaker:ubuntu_1604_single_node

acceptance-1-8-ubuntu: export KUBERNETES_VERSION = $(VERSION_1_8)
acceptance-1-8-ubuntu: bundle_install
	bundle exec rake beaker:ubuntu_1604_single_node

acceptance-1-9-ubuntu: export KUBERNETES_VERSION = $(VERSION_1_9)
acceptance-1-9-ubuntu: bundle_install
	bundle exec rake beaker:ubuntu_1604_single_node

acceptance-1-10-ubuntu: export KUBERNETES_VERSION = $(VERSION_1_10)
acceptance-1-10-ubuntu: bundle_install
	bundle exec rake beaker:ubuntu_1604_single_node

acceptance-1-11-ubuntu: export KUBERNETES_VERSION = $(VERSION_1_11)
acceptance-1-11-ubuntu: bundle_install
	bundle exec rake beaker:ubuntu_1604_single_node

acceptance-1-12-ubuntu: export KUBERNETES_VERSION = $(VERSION_1_12)
acceptance-1-12-ubuntu: bundle_install
	bundle exec rake beaker:ubuntu_1604_single_node

acceptance-1-13-ubuntu: export KUBERNETES_VERSION = $(VERSION_1_13)
acceptance-1-13-ubuntu: bundle_install
	bundle exec rake beaker:ubuntu_1604_single_node
