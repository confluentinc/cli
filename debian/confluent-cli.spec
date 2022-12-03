Name: confluent-cli
Version: %{__version}
Release: %{__release}%{?dist}
Summary: CLI for Confluent Cloud and Confluent Platform
Group: Applications/Internet
License: Confluent License Agreement
Source0: confluent-cli-%{version}.tar.gz
URL: http://confluent.io
Vendor: Confluent, Inc.
Packager: Confluent Packaging <packages@confluent.io>

BuildRoot: %{_tmppath}/%{name}-%{version}-root

%description
The Confluent CLI helps you manage your Confluent Cloud and Confluent Platform deployments, right from the terminal.

%global debug_package %{nil}

%prep

%setup

%build

%install
rm -rf %{buildroot}
DESTDIR=%{buildroot} make install

%files
%defattr(-,root,root)
/usr/bin/*
%doc /usr/share/doc/cli/

%clean
rm -rf %{buildroot}

%changelog
* Fri Jul 24 2020 Confluent Packaging <packages@confluent.io>
- Initial import
