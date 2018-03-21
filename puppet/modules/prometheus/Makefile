BUNDLE_DIR ?= .bundle

export PATH := $(CURDIR)/spec/fixtures/bin:$(PATH)

verify: bundle_install binary_install
	which promtool
	bundle exec rake test

bundle_install:
	bundle install --path $(BUNDLE_DIR)

binary_install: spec/fixtures/bin/promtool

spec/fixtures/bin/promtool:
	mkdir -p spec/fixtures/bin/
	curl -Lo spec/fixtures/prometheus.tar.gz https://github.com/prometheus/prometheus/releases/download/v2.2.1/prometheus-2.2.1.linux-amd64.tar.gz
	echo "e12917b25b32980daee0e9cf879d9ec197e2893924bd1574604eb0f550034d46  spec/fixtures/prometheus.tar.gz" | sha256sum -c
	tar xvfz spec/fixtures/prometheus.tar.gz -C spec/fixtures/bin --strip-components=1 prometheus-2.2.1.linux-amd64/promtool

acceptance: bundle_install
	bundle exec rake beaker:default
