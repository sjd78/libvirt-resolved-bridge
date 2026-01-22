package main
import "testing"

func TestParseNetworkXML(t *testing.T) {
	xmlStr := `<network><name>minikube</name><bridge name="mkbr0"/><domain name="mk.local"/><ip address="192.168.1.1"/></network>`
	cfg, err := parseNetworkXML([]byte(xmlStr))
	if err != nil { t.Fatal(err) }
	if cfg.Bridge.Name != "mkbr0" { t.Errorf("Expected mkbr0, got %s", cfg.Bridge.Name) }
}
