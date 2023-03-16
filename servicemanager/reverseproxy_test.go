package servicemanager

import (
	"testing"
)

func Test_buildRoutingTable(t *testing.T) {
	service := Service{Id: "Foo", DefaultPort: 8080, ProxyPaths: []string{"/path1", "/path2"}}
	result := buildRoutingTable(map[string]Service{"Foo": service})
	if v, ok := result["/path1"]; !ok || v != "localhost:8080" {
		t.Errorf("Routes /path1 did not have expected value. Expected value was %s actual value was %s", "localhost:8080", v)
	}
	if v, ok := result["/path2"]; !ok || v != "localhost:8080" {
		t.Errorf("Routes /path2 did not have expected value. Expected value was %s actual value was %s", "localhost:8080", v)
	}
}
