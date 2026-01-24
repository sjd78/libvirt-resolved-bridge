package main

import (
	"encoding/xml"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/godbus/dbus/v5"
)

type NetworkXML struct {
	Name   string `xml:"name"`
	Domain struct {
		Name string `xml:"name,attr"`
	} `xml:"domain"`
	Bridge struct {
		Name string `xml:"name,attr"`
	} `xml:"bridge"`
	IP struct {
		Address string `xml:"address,attr"`
	} `xml:"ip"`
}

// DBus struct types for systemd-resolved
// SetLinkDNS expects: a(iay) - array of (int32 family, []byte address)
type dnsEntry struct {
	Family  int32
	Address []byte
}

// SetLinkDomains expects: a(sb) - array of (string domain, bool routing_only)
type domainEntry struct {
	Domain      string
	RoutingOnly bool
}

const (
	libvirtDest    = "org.libvirt"
	resolvedDest   = "org.freedesktop.resolve1"
	resolvedObject = "/org/freedesktop/resolve1"
)

func main() {
	conn, err := dbus.SystemBus()
	if err != nil {
		log.Fatalf("Failed to connect to system bus: %v", err)
	}
	defer conn.Close()

	log.Println("Connected to system bus")

	// Ping libvirt-dbus to ensure it's activated and ready
	libvirtObj := conn.Object("org.libvirt", "/org/libvirt/QEMU")
	var introspectXML string
	err = libvirtObj.Call("org.freedesktop.DBus.Introspectable.Introspect", 0).Store(&introspectXML)
	if err != nil {
		log.Printf("Warning: Could not introspect libvirt-dbus (may not be ready): %v", err)
	} else {
		log.Println("libvirt-dbus is responding")
	}

	// Listen for NetworkEvent signals from org.libvirt.Connect interface
	if err := conn.AddMatchSignal(
		dbus.WithMatchSender("org.libvirt"),
		dbus.WithMatchObjectPath("/org/libvirt/QEMU"),
		dbus.WithMatchInterface("org.libvirt.Connect"),
		dbus.WithMatchMember("NetworkEvent"),
	); err != nil {
		log.Fatalf("Failed to add match signal: %v", err)
	}
	log.Println("Match signal registered successfully")

	signals := make(chan *dbus.Signal, 10)
	conn.Signal(signals)

	log.Println("Libvirt-Resolved Bridge started. Monitoring Libvirt networks...")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case sig := <-signals:
			handleSignal(conn, sig)
		case <-sigChan:
			log.Println("Shutting down...")
			return
		}
	}
}

func handleSignal(conn *dbus.Conn, sig *dbus.Signal) {
	// NetworkEvent signal: (o network, i event)
	if sig.Name != "org.libvirt.Connect.NetworkEvent" {
		return
	}

	if len(sig.Body) < 2 {
		return
	}

	netPath, ok1 := sig.Body[0].(dbus.ObjectPath)
	event, ok2 := sig.Body[1].(int32)
	if !ok1 || !ok2 {
		log.Printf("Unexpected signal body types: %T, %T", sig.Body[0], sig.Body[1])
		return
	}

	// https://libvirt.org/html/libvirt-libvirt-network.html#virNetworkEventLifecycleType
	switch event {
	case 0:
		log.Printf("Network '%s' detected as DEFINED", netPath)
	case 1:
		log.Printf("Network '%s' detected as UNDEFINED", netPath)
	case 2: // Started (VIR_NETWORK_EVENT_STARTED)
		log.Printf("Network '%s' detected as STARTED", netPath)
		processNetwork(conn, netPath)
	case 3: // Stopped (VIR_NETWORK_EVENT_STOPPED)
		log.Printf("Network '%s' detected as STOPPED", netPath)
	}
}

func parseNetworkXML(data []byte) (*NetworkXML, error) {
	var config NetworkXML
	if err := xml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func processNetwork(conn *dbus.Conn, netPath dbus.ObjectPath) {
	netObj := conn.Object(libvirtDest, netPath)

	var xmlData string
	err := netObj.Call("org.libvirt.Network.GetXMLDesc", 0, uint32(0)).Store(&xmlData)
	if err != nil {
		log.Printf("Could not get XML for network %s: %v", netPath, err)
		return
	}

	config, err := parseNetworkXML([]byte(xmlData))
	if err != nil {
		log.Printf("Failed to parse XML for %s: %v", netPath, err)
		return
	}

	bridge := config.Bridge.Name
	domain := config.Domain.Name
	dnsIP := config.IP.Address

	log.Printf("Network config - bridge: %s, domain: %s, dns: %s", bridge, domain, dnsIP)

	if bridge == "" || (domain == "" && dnsIP == "") {
		log.Printf("No bridge or domain/DNS IP found for %s", netPath)
		return
	}

	applyToResolved(conn, bridge, domain, dnsIP)
}

func applyToResolved(conn *dbus.Conn, ifName, domain, dnsIP string) {
	iface, err := net.InterfaceByName(ifName)
	if err != nil {
		log.Printf("Could not find interface %s: %v", ifName, err)
		return
	}

	resObj := conn.Object(resolvedDest, resolvedObject)
	ifIndex := int32(iface.Index)

	if dnsIP != "" {
		ip := net.ParseIP(dnsIP).To4()
		if ip != nil {
			log.Printf("Setting DNS for interface %s (idx %d) to %s", ifName, ifIndex, ip.String())
			// AF_INET = 2 for IPv4
			dnsEntries := []dnsEntry{{Family: 2, Address: []byte(ip)}}
			call := resObj.Call("org.freedesktop.resolve1.Manager.SetLinkDNS", 0, ifIndex, dnsEntries)
			if call.Err != nil {
				log.Printf("Failed to set DNS: %v", call.Err)
			}
		}
	}

	if domain != "" {
		log.Printf("Setting domain for interface %s (idx %d) to %s", ifName, ifIndex, domain)
		// routing_only=true means this domain is only for routing, not search
		domainEntries := []domainEntry{{Domain: domain, RoutingOnly: true}}
		call := resObj.Call("org.freedesktop.resolve1.Manager.SetLinkDomains", 0, ifIndex, domainEntries)
		if call.Err != nil {
			log.Printf("Failed to set domains: %v", call.Err)
		}
	}
}
