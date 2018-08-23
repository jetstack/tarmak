BUNDLE_DIR ?= .bundle

verify: bundle_install
	bundle exec rake test

bundle_install:
	bundle install --path $(BUNDLE_DIR)
