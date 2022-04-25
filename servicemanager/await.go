package servicemanager

import (
	"fmt"
	"time"
)

// TODO: have this return the results so we can decide how to format it elsewhere
//       this way we can repsect --no-progress etc
func (sm *ServiceManager) Await(services []ServiceAndVersion, timeout int) {

	// track statuses in a map
	health := map[string]bool{}
	for _, s := range services {
		health[s.service] = false
	}

	// poll statues to see whats running
	t := 0
	healthy := 0
	for t < timeout && healthy < len(health) {
		healthy = 0 // reset health service count
		t++
		statuses := sm.findStatuses()
		for _, status := range statuses {
			if _, ok := health[status.service]; ok {
				if status.health == PASS {
					health[status.service] = true
					healthy++
				}
			}
		}
		time.Sleep(time.Second)
	}

	// print results
	if healthy == len(health) {
		fmt.Println("All services started ok.")
	}
	for k, isHealthy := range health {
		if !isHealthy {
			fmt.Printf("%s failed to start.\n", k)
		}
	}

}
