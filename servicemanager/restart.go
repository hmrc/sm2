package servicemanager

import (
	"fmt"
)

// restarts a service using the previous configuration
func (sm *ServiceManager) Restart(sv ServiceAndVersion) error {

	// verify its a real service
	service, ok := sm.Services[sv.service]
	if !ok {
		return fmt.Errorf("%s is not a service", sv.service)
	}

	// read state file
	state, err := sm.Ledger.LoadStateFile(sv.service)
	if err != nil {
		return err
	}

	// read install file
	install, err := sm.Ledger.LoadInstallFile(sv.service)
	if err != nil {
		return err
	}

	// check its ok
	if !verifyInstall(install, sv.service, sv.version, false) {
		return fmt.Errorf("%s %s is not installed", sv.service, sv.version)
	}

	// stop the current service
	if err := sm.StopService(sv.service); err != nil {
		return err
	}

	// start a new instance
	newstate, err := run(service, install, state.Args, state.Port)

	// save the new pid
	installDir, _ := sm.findInstallDirOfService(sv.service)
	return sm.Ledger.SaveStateFile(installDir, newstate)
}
