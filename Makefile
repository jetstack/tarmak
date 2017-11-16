BUNDLE_DIR ?= .bundle

verify: bundle_install
	bundle exec rake test

bundle_install:
	bundle install --path $(BUNDLE_DIR)

acecptance: acceptance-1-7-centos

acceptance-1-8-centos: export KUBERNETES_VERSION = 1.8.3
acceptance-1-8-centos: bundle_install
	bundle exec rake beaker:default

acceptance-1-7-centos: export KUBERNETES_VERSION = 1.7.10
acceptance-1-7-centos: bundle_install
	bundle exec rake beaker:default

acceptance-1-6-centos: export KUBERNETES_VERSION = 1.6.12
acceptance-1-6-centos: bundle_install
	bundle exec rake beaker:default

acceptance-1-5-centos: export KUBERNETES_VERSION = 1.5.8
acceptance-1-5-centos: export KUBERNETES_AUTHORIZATION_MODE = ['RBAC']
acceptance-1-5-centos: bundle_install
	bundle exec rake beaker:default

acceptance-1-7-ubuntu: export KUBERNETES_VERSION = 1.7.10
acceptance-1-7-ubuntu: bundle_install
	bundle exec rake beaker:ubuntu_1604_single_node
