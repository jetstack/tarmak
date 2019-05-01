BUNDLE_DIR ?= .bundle

verify: bundle_install
	bundle exec rake test

bundle_install:
	bundle install --path $(BUNDLE_DIR)

acceptance: bundle_install
	bundle exec rake beaker:default

acceptance-single-node: acceptance

# provide backwards compatiblity
acceptance_three_node: acceptance-three-node

acceptance-three-node: bundle_install
	bundle exec rake beaker:3n-cluster
