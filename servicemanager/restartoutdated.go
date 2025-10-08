package servicemanager

import (
	"fmt"
)

// restarts services running outdated versions
func (sm *ServiceManager) RestartOutdated() {
	outdatedServices := []ServiceAndVersion{}

	// check for a newer version for each service running
	for _, status := range sm.findStatuses() {
		serviceAndVersion := ServiceAndVersion{status.service, "", ""}
		_, _, LatestVersion, _ := whatVersionToRun(
			sm.Services[status.service],
			serviceAndVersion,
			false,
			sm.GetLatestVersions)

		// if there is a newer version, stop the service and save it for later
		if status.version != LatestVersion {
			sm.StopService(status.service)
			outdatedServices = append(outdatedServices, serviceAndVersion)
		}
	}

	// start stopped services with latest versions
	if len(outdatedServices) > 0 {
		fmt.Println()
		sm.asyncStart(outdatedServices)
	} else {
		fmt.Println("All services are running latest versions")
	}
}
