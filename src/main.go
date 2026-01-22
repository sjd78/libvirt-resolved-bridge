package main

import (
	"encoding/xml"
	"fmt"
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

const (
	libvirtDest    = "org.libvirt.virt"
	resolvedDest   = "org.freedesktop.resolve1"
	resolvedObject = "/org/freedesktop/resolve1"
)

func main() {
	conn, err := dbus.SystemBus()
	if err != nil {
		log.Fatalf("Failed to connect to system bus: %v", err)
	}
	defer conn.Close()

	rule := "type='signal',sender='org.libvirt.virt',interface='org.libvirt.Network'"
	conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, rule)

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
	if len(sig.Body) < 2 {
		return
	}
	netName, ok1 := sig.Body[0].(string)
	event, ok2 := sig.Body[1].(int32)
	if !ok1 || !ok2 {
		return
	}

	switch event {
	case 0: // Started
		log.Printf("Network '%s' detected as STARTED", netName)
		processNetwork(conn, netName)
	case 1: // Stopped
		log.Printf("Network '%s' detected as STOPPED", netName)
	}
}

func parseNetworkXML(data []byte) (*NetworkXML, error) {
	var config NetworkXML
	if err := xml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func processNetwork(conn *dbus.Conn, netName string) {
	objPath := dbus.ObjectPath(fmt.Sprintf("/org/libvirt/virt/network/%s", netName))
	netObj := conn.Object(libvirtDest, objPath)

	var xmlData string
	err := netObj.Call("org.libvirt.Network.GetXMLDesc", 0, uint32(0)).Store(&xmlData)
	if err != nil {
		log.Printf("Could not get XML for network %s: %v", netName, err)
		return
	}

	config, err := parseNetworkXML([]byte(xmlData))
	if err != nil {
		log.Printf("Failed to parse XML for %s: %v", netName, err)
		return
	}

	bridge := config.Bridge.Name
	domain := config.Domain.Name
	dnsIP := config.IP.Address

	if bridge == "" || (domain == "" && dnsIP == "") {
		log.Printf("No bridge or domain/DNS IP found for %s", netName)
		return
	}

	applyToResolved(conn, bridge, domain, dnsIP)
}

func applyToResolved(conn *dbus.Conn, ifName, domain, dnsIP string) {
	iface, err := net.InterfaceByName(ifName)
	if err != nil {
		return
	}

	resObj := conn.Object(resolvedDest, resolvedObject)
	ifIndex := int32(iface.Index)

	if dnsIP != "" {
		ip := net.ParseIP(dnsIP).To4()
		if ip != nil {
			dnsEntry := []interface{}{int32(2), []byte(ip)}
			resObj.Call("org.freedesktop.resolve1.Manager.SetLinkDNS", 0, ifIndex, []interface{}{dnsEntry})
		}
	}

	if domain != "" {
		domainEntry := []interface{}{domain, true}
		resObj.Call("org.freedesktop.resolve1.Manager.SetLinkDomains", 0, ifIndex, []interface{}{domainEntry})
	}
}
