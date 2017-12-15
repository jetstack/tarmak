# Copyright Jetstack Ltd. See LICENSE for details.
# Minimal makefile for Sphinx documentation
#

# You can set these variables from the command line.
VENV_PATH    ?= venv
SPHINXOPTS    = -W
SPHINXBUILD   = $(VENV_PATH)/bin/python -msphinx
SPHINXPROJ    = Tarmak
SOURCEDIR     = .
BUILDDIR      = _build

LC_ALL = C.UTF-8
LANG = C.UTF-8
export LANG LC_ALL

# Put it first so that "make" without argument is like "make help".
help:
	make " help"

.PHONY: help Makefile clean

# Catch-all target: route all unknown targets to Sphinx using the new
# "make mode" option.  $(O) is meant as a shortcut for $(SPHINXOPTS).
%: Makefile .venv
	@$(SPHINXBUILD) -M $@ "$(SOURCEDIR)" "$(BUILDDIR)" $(SPHINXOPTS) $(O)

.venv:
	virtualenv -p $(shell which python3) $(VENV_PATH)
	$(VENV_PATH)/bin/pip install -r requirements.txt
	touch .venv

clean:
	rm -rf .venv "$(BUILDDIR)"

copy_artifacts:
	mkdir -p ../_artifacts
	cp -a $(BUILDDIR)/html ../_artifacts/docs

multiversion_build: .venv
	$(VENV_PATH)/bin/sphinx-versioning build -r master -w "^(master|release-\d+\.\d+)$$" -W "none" ./docs ./_build

multiversion_push: CLOUDSDK_CONFIG = $(CURDIR)/.gcloud
multiversion_push:
	mkdir -p $(CLOUDSDK_CONFIG)
	chmod 644 $(CLOUDSDK_CONFIG)
	gcloud auth activate-service-account --key-file $(DOCS_GOOGLE_APPLICATION_CREDENTIALS)
	gsutil -m rsync -rd _build/ gs://docs.tarmak.io/
	rm -rf $(CLOUDSDK_CONFIG)
