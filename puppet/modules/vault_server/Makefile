BUNDLE_DIR ?= .bundle

verify: bundle_install
	bundle exec rake test

bundle_install:
	bundle install --path $(BUNDLE_DIR)

acceptance: accetance-centos

acceptance-centos: bundle_install
	bundle exec rake beaker:default

acceptance-ubuntu: bundle_install
	bundle exec rake beaker:ubuntu_1604_single_node
