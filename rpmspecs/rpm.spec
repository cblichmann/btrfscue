Name:	        btrfscue
Version:	0.0.0
Release:	1%{?dist}
Summary:	btrfscue
License:	BSD-2-Clause
URL:		https://github.com/cblichmann/btrfscue
Source0:  ../
BuildArch:      x86_64

#Requires:

BuildRequires:	go

%description
Btrfs Recucovery tool

%install
echo "BUILD_ROOT = $RPM_BUILD_ROOT"
mkdir -p $RPM_BUILD_ROOT/usr/local/bin
cp ../../bin/* $RPM_BUILD_ROOT/usr/local/bin
exit

%files
%attr(0755, root, root) /usr/local/bin/*
#%doc
#%changelog
%clean
rm -rf $RPM_BUILD_ROOT/usr/local/bin
