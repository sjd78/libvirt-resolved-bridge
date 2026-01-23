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

### Installation
1. Compile the binary: `go build -o libvirt-resolved-bridge ./src`
2. Move it to `/usr/local/bin/`
3. Install the systemd unit provided in `libvirt-resolved-bridge.service`.
