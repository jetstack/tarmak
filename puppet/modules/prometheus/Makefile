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
	curl -Lo spec/fixtures/prometheus.tar.gz https://github.com/prometheus/prometheus/releases/download/v2.3.2/prometheus-2.3.2.linux-amd64.tar.gz
	echo "351931fe9bb252849b7d37183099047fbe6d7b79dcba013fb6ae19cc1bbd8552  spec/fixtures/prometheus.tar.gz" | sha256sum -c
	tar xvfz spec/fixtures/prometheus.tar.gz -C spec/fixtures/bin --strip-components=1 prometheus-2.3.2.linux-amd64/promtool

acceptance: bundle_install
	bundle exec rake beaker:default
