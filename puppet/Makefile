# Copyright Jetstack Ltd. See LICENSE for details.
BUNDLE_DIR ?= .bundle

verify: bundle_install
	bundle exec rake spec

bundle_install:
	bundle install --path $(BUNDLE_DIR)
