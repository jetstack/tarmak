# Copyright Jetstack Ltd. See LICENSE for details.
# Minimal makefile for Sphinx documentation
#

# You can set these variables from the command line.
SPHINXOPTS    = -W
SPHINXBUILD   = python -msphinx
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
	@. venv/bin/activate && $(SPHINXBUILD) -M $@ "$(SOURCEDIR)" "$(BUILDDIR)" $(SPHINXOPTS) $(O)

.venv:
	virtualenv venv
	. venv/bin/activate && pip install -r requirements.txt
	touch .venv

clean:
	rm -rf venv .venv "$(BUILDDIR)"
