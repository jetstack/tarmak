BUNDLE_DIR ?= .bundle

verify: bundle_install
	bundle exec rake test

bundle_install:
	bundle install --path $(BUNDLE_DIR)

acceptance: bundle_install
	bundle exec rake beaker:default

acceptance_three_node: bundle_install
	bundle exec rake beaker:3n-cluster
