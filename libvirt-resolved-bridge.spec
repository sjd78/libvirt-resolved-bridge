Name:           libvirt-resolved-bridge
Version:        0.1.0
Release:        1%{?dist}
Summary:        Bridge Libvirt network configurations to systemd-resolved

# Go binaries don't produce debuginfo
%global debug_package %{nil}

License:        Apache-2.0
URL:            https://github.com/sjd78/%{name}
Source0:        %{name}-%{version}.tar.gz

BuildRequires:  make
BuildRequires:  git
BuildRequires:  golang >= 1.21
BuildRequires:  systemd-rpm-macros

Requires:       libvirt-dbus
Requires:       systemd-resolved

%description
A Go daemon that listens for libvirt network events via DBus and automatically
configures systemd-resolved with the appropriate DNS settings for each virtual
network. This enables seamless DNS resolution for VMs on libvirt networks.

%prep
%autosetup -n %{name}-%{version}

%build
%make_build

%install
%make_install

%post
%systemd_post %{name}.service

%preun
%systemd_preun %{name}.service

%postun
%systemd_postun_with_restart %{name}.service

%files
%license LICENSE
%doc README.md
%{_bindir}/%{name}
%{_unitdir}/%{name}.service

%changelog
* Thu Jan 22 2026 Scott Dickerson <sdickers@redhat.com> - 0.1.0-1
- Initial package release
