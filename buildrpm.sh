#!/bin/sh
rpmbuild -bb rpmbuild/SPECS/rpm.spec --define "_topdir $(pwd)"
