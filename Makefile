verify: bundle_install
	bundle exec rake test

bundle_install:
	bundle install --path .bundle

acceptance: bundle_install
	bundle exec rake beaker:default
