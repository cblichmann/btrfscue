#!/usr/bin/env make
#
# Copyright btrfscue authors
# SPDX-License-Identifier: BSD-2-Clause
#
# Makefile for POSIX compatible systems

# Source Configuration
version = 0.7
c_year ?= $(shell date +%Y)
go_package = blichmann.eu/code/btrfscue
go_programs = btrfscue
source_only_tgz = ../btrfscue_$(version).orig.tar.xz

# Directories
this_dir := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
bin_dir := $(this_dir)/bin

export GOBIN := $(bin_dir)

binaries := $(addprefix $(bin_dir)/,$(go_programs))
coverage := c.out
sources := $(wildcard $(shell (unset GOPATH; go list -f '{{.Dir}}/*.go' ./...)))

.PHONY: all
all: $(binaries)

.PHONY: clean
clean:
	@echo "  [Clean]     Removing build artifacts"
	@rm -f $(binaries) $(coverage) || true
	@rmdir "$(bin_dir)" || true

$(binaries): $(sources)
	@echo "  [Build]     $@"
	@(unset GOPATH; go install -tags "$(TAGS)" $(go_package)/...)

.PHONY: test
test:
	@echo "  [Test]"
	@(unset GOPATH; go test $(COVER_ARGS) ./...)

$(coverage): COVER_ARGS=-coverprofile=$(coverage)
$(coverage): test
	@echo "  [Coverage]  $@"

.PHONY: coverage
coverage: $(coverage)

.PHONY: show_coverage
show_coverage:
	@go tool cover -html=c.out

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
		[ -f $$i ] && sed -i \
			-e 's/\(btrfscue version\) [0-9]\+\.[0-9]\+/\1 $(version)/' \
			-e 's/\(Copyright (c)[0-9]\+\)-[0-9]\+/\1-$(c_year)/' \
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
