package servicemanager

import (
	"fmt"
	"sort"
)

type duplicatePort struct {
	Port     int
	ServiceA string
	ServiceB string
}

// Prints a list of ports that are used by more than one service
func (sm *ServiceManager) checkPorts() {

	duplicates := findDuplicatePorts(sm.Services)

	sort.Slice(duplicates, func(i, j int) bool {
		return duplicates[i].Port < duplicates[j].Port
	})

	for _, d := range duplicates {
		fmt.Printf("Duplicate port found: %d in services: %s and %s\n", d.Port, d.ServiceA, d.ServiceB)
	}

}

func findDuplicatePorts(services map[string]Service) []duplicatePort {

	portsSeen := make(map[int]string, len(services))
	duplicates := []duplicatePort{}

	for _, service := range services {

		if service.DefaultPort == 0 {
			// skip services without a port
			continue
		}

		if id, ok := portsSeen[service.DefaultPort]; ok {
			duplicates = append(duplicates, duplicatePort{service.DefaultPort, id, service.Id})
		}

		portsSeen[service.DefaultPort] = service.Id
	}

	return duplicates
}
