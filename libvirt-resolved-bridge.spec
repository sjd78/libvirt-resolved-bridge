Name:           libvirt-resolved-bridge
Version:        0.1.0
Release:        1%{?dist}
Summary:        Bridge Libvirt network configurations to systemd-resolved
License:        ASL 2.0
URL:            https://github.com/sjd78/%{name}
Source0:        %{name}-%{version}.tar.gz
BuildRequires:  golang, systemd-rpm-macros
Requires:       libvirt-dbus, systemd-resolved

%description
A Go daemon that listens for libvirt network events and configures systemd-resolved.

%prep
%setup -q

%build
go build -v -o %{name} ./src

%install
install -D -m 0755 %{name} %{buildroot}%{_bindir}/%{name}
install -D -m 0644 %{name}.service %{buildroot}%{_unitdir}/%{name}.service

%files
%license LICENSE
%doc README.md
%{_bindir}/%{name}
%{_unitdir}/%{name}.service
