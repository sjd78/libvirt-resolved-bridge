![copr status](https://copr.fedorainfracloud.org/coprs/sdickers/libvirt-resolved-bridge/package/libvirt-resolved-bridge/status_image/last_build.png)
[![CI](https://github.com/sjd78/libvirt-resolved-bridge/actions/workflows/ci.yml/badge.svg?branch=main&event=push)](https://github.com/sjd78/libvirt-resolved-bridge/actions/workflows/ci.yml)

# libvirt-resolved-bridge

A Go daemon that bridges Libvirt network configurations to `systemd-resolved`.

### How it works
1. It listens for `org.libvirt.Network` signals via DBus.
2. When a network is started, it fetches the network XML from Libvirt.
3. It parses the `<domain name="...">` and `<ip address="...">` fields.
4. It calls `org.freedesktop.resolve1` to configure the DNS routing for that specific bridge interface.

### Requirements
- Fedora Linux (or any system with `systemd-resolved`)
- `libvirt-dbus` installed and running (`sudo dnf install libvirt-dbus`)
- Go 1.21+

### Build and installation

#### To manually build and install the service:
```sh
make build
sudo make install
sudo systemctl daemon-reload
sudo systemctl enable --now libvirt-resolved-bridge
```

#### Install from the copr tracking the main branch in this repo:
```sh
sudo dnf copr enable sdickers/libvirt-resolved-bridge
sudo dnf install libvirt-resolved-bridge
```
