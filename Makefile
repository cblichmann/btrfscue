#!/usr/bin/env make
#
# btrfscue version 0.3
# Copyright (c)2015,2016 Christian Blichmann
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
go_package = blichmann.eu/code/btrfscue
go_programs = btrfscue

# Directories
this_dir := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
bin_dir := $(this_dir)/bin
pkg_src := $(this_dir)/src/$(go_package)
third_party_dir := $(abspath $(this_dir)/../third_party)

binaries := $(addprefix $(bin_dir)/,$(go_programs))
sources := $(wildcard $(shell go list -f '{{.Dir}}/*.go' ./...))
project_go_path := $(third_party_dir)/go:$(this_dir)

.PHONY: all
all: $(binaries)

.PHONY: env
env:
#	# Use like this: eval $(make env)
	@echo "export GOPATH=$(project_go_path)"

.PHONY: clean
clean:
	@echo "  [Clean]     Removing build artifacts"
	@for i in bin pkg src; do rm -rf "$(this_dir)/$$i"; done

$(binaries): export GOPATH = $(project_go_path)
$(binaries): $(sources)
	@echo "  [Build]     $@"
	@mkdir -p "$(dir $(pkg_src))"
	@ln -sft "$(dir $(pkg_src))" "$(this_dir)"
	@go install -tags '$(TAGS)' $(go_package)
