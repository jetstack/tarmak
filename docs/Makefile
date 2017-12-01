# Copyright Jetstack Ltd. See LICENSE for details.
# Minimal makefile for Sphinx documentation
#

# You can set these variables from the command line.
SPHINXOPTS    = -W
SPHINXBUILD   = venv/bin/python -msphinx
SPHINXPROJ    = Tarmak
SOURCEDIR     = .
BUILDDIR      = _build

# Put it first so that "make" without argument is like "make help".
help:
	make " help"

.PHONY: help Makefile clean

# Catch-all target: route all unknown targets to Sphinx using the new
# "make mode" option.  $(O) is meant as a shortcut for $(SPHINXOPTS).
%: Makefile .venv
	@$(SPHINXBUILD) -M $@ "$(SOURCEDIR)" "$(BUILDDIR)" $(SPHINXOPTS) $(O)

.venv:
	virtualenv -p $(shell which python3) venv
	venv/bin/pip install -r requirements.txt
	touch .venv

clean:
	rm -rf venv .venv "$(BUILDDIR)"

multiversion_build:
	venv/bin/sphinx-versioning build -r master -w "^(master|release-\d+\.\d+)$$" -W "none" ./docs ./_build

multiversion_push:
	gsutil -m rsync -rd _build/ gs://docs.tarmak.io/
