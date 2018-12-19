#!/usr/bin/env make
#
# btrfscue version 0.5
# Copyright (c)2011-2018 Christian Blichmann
#
# Makefile for POSIX compatible systems
#
# Redistribution and use in source and binary forms, with or without
# modification, are permitted provided that the following conditions are met:
#     * Redistributions of source code must retain the above copyright
#       notice, this list of conditions and the following disclaimer.
#     * Redistributions in binary form must reproduce the above copyright
#       notice, this list of conditions and the following disclaimer in the
#       documentation and/or other materials provided with the distribution.
#
# THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
# AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
# IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
# ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
# LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
# CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
# SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
# INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
# CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
# ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
# POSSIBILITY OF SUCH DAMAGE.

# Source Configuration
version = 0.5
go_package = blichmann.eu/code/btrfscue
go_programs = btrfscue
source_only_tgz = ../btrfscue_$(version).orig.tar.xz

# Directories
this_dir := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
bin_dir := $(this_dir)/bin
pkg_src := $(this_dir)/src/$(go_package)

binaries := $(addprefix $(bin_dir)/,$(go_programs))
sources := $(wildcard $(shell go list -f '{{.Dir}}/*.go' ./...))
GOPATH := $(this_dir)

.PHONY: all
all: $(binaries)

.PHONY: env
env:
#	# Use like this: eval $(make env)
	@echo "export GOPATH=$(GOPATH)"

.PHONY: clean
clean:
	@echo "  [Clean]     Removing build artifacts"
	@for i in bin pkg src; do rm -rf "$(this_dir)/$$i"; done

$(binaries): $(sources)
	@echo "  [Build]     $@"
	@mkdir -p "$(dir $(pkg_src))"
	@test -h "$(pkg_src)" || ln -s "$(this_dir)" "$(pkg_src)"
	@cd "$(pkg_src)" && go install -tags "$(TAGS)" $(go_package)

.PHONY: test
test: $(binaries)
	@echo "  [Test]"
	@cd "$(pkg_src)" && go test ./...

$(source_only_tgz): clean
	@echo "  [Archive]   $@"
	@tar -C "$(this_dir)" -caf "$@" \
		--transform=s,^,btrfscue-$(version)/, \
		--exclude=.git/* --exclude=.git \
		--exclude=__tmp/* --exclude=__tmp \
		--exclude=debian/* --exclude=debian \
		"--exclude=$@" \
		--exclude-vcs-ignores \
		.??* *

.PHONY: updatesourcemeta
updatesourcemeta:
	@echo "  [Update]    Version and copyright"
	@for i in \
		$(sources) \
		$(this_dir)/debian/copyright \
		$(this_dir)/debian/rules \
		$(this_dir)/man/*.[1-9] \
		$(this_dir)/LICENSE \
		$(this_dir)/Makefile \
		$(this_dir)/README.md; \
	do \
		sed -i \
			-e 's/\(btrfscue version\) [0-9]\+\.[0-9]\+/\1 $(version)/' \
			-e 's/\(Copyright (c)[0-9]\+\)-[0-9]\+/\1-$(shell date +%Y)/' \
			$$i; \
	done

# Create a source tarball without the debian/ subdirectory
.PHONY: debsource
debsource: $(source_only_tgz)

# debuild signs the package iff DEBFULLNAME, DEBEMAIL and DEB_SIGN_KEYID are
# set. Note that if the GPG key includes an alias, it must match the latest
# entry in debian/changelog.
deb: debsource $(binaries)
	@echo "  [Debuild]   Building package"
	@debuild

.PHONY: debclean
debclean: clean
	@echo "  [Deb-Clean] Removing artifacts"
	@debuild -- clean
