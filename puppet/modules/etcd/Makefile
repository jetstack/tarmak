BUNDLE_DIR ?= .bundle

verify: bundle_install
	bundle exec rake test

bundle_install:
	bundle install --path $(BUNDLE_DIR)

acceptance-single-node: bundle_install
	bundle exec rake beaker:default

acceptance-three-node: bundle_install
	bundle exec rake beaker:3n-cluster

# provide backwards compatiblity
acceptance: acceptance-single-node
acceptance_three_node: acceptance-three-node

