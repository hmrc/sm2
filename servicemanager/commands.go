package servicemanager

import (
	"fmt"
	"os"
	"regexp"
	"sm2/cli"
	"sm2/version"
)

type ServiceAndVersion struct {
	service      string
	version      string
	scalaVersion string
}

var serviceAndVersionRegex *regexp.Regexp = regexp.MustCompile(`(.*?)(_(2\.\d{2}|3))?(:(.*))?$`)

func parseServiceAndVersion(serviceDescriptor string) ServiceAndVersion {
	matches := serviceAndVersionRegex.FindStringSubmatch(serviceDescriptor)

	if matches == nil {
		return ServiceAndVersion{serviceDescriptor, "", ""}
	} else {
		service := matches[1]
		scalaVersion := matches[3]
		version := matches[5]
		return ServiceAndVersion{service, version, scalaVersion}
	}
}

func (sm *ServiceManager) Run() {

	var err error

	if sm.Commands.UpdateConfig {
		err := updateConfig(sm.Config.ConfigDir)
		if err != nil {
			fmt.Println(err)
			fmt.Println("Continuing with current config...")
		}
		err = sm.LoadConfig()
		if err != nil {
			fmt.Print(err)
			os.Exit(1)
		}
	}

	if sm.Commands.Status || sm.Commands.StatusShort {
		// prints table of running services
		sm.PrintStatus()
	} else if sm.Commands.Prune {
		// cleans up state files for services with a status of FAIL
		sm.cleanupFailedServices()
	} else if sm.Commands.Start {
		// starts service(s) or profile(s)
		services := sm.requestedServicesAndProfiles()
		sm.asyncStart(services)
	} else if sm.Commands.Stop {
		// stops a specific service or profile
		services := sm.requestedServicesAndProfiles()
		for _, s := range services {
			err = sm.StopService(s.service)
		}
	} else if sm.Commands.StopAll {
		// stops all managed services
		sm.StopAll()
	} else if sm.Commands.Restart {
		// restarts service(s) or profile(s)
		services := sm.requestedServicesAndProfiles()
		failed := []ServiceAndVersion{}
		for _, s := range services {
			if err := sm.Restart(s); err != nil {
				failed = append(failed, s)
			}
		}
		// try and start the failed services (which are probably just not running)
		if len(failed) > 0 {
			sm.asyncStart(failed)
		}
	} else if sm.Commands.Ports {
		// prints all port numbers to stdout
		sm.ListPorts()
	} else if sm.Commands.CheckPorts {
		sm.checkPorts()
	} else if sm.Commands.Search != "" {
		// regex search of services and profiles
		sm.ListServices(sm.Commands.Search, sm.Commands.FormatPlain)
	} else if sm.Commands.List {
		// alias for search everything
		sm.ListServices(".", sm.Commands.FormatPlain)
	} else if sm.Commands.Logs != "" {
		// dumps stdout.log to stdout
		sm.PrintLogsForService(sm.Commands.Logs)
	} else if sm.Commands.ReverseProxy {
		// starts a reverse proxy for frontend services
		sm.StartProxy()
	} else if sm.Commands.Offline {
		// used by itself, offline will list available services
		sm.ListServicesAvailableOffline()
	} else if sm.Commands.Diagnostic {
		// checks if system can run sm2
		RunDiagnostics(sm.Config)
	} else if sm.Commands.Debug != "" {
		// `--debug SERVICE` dumps as much info as it can find about the service
		sm.showDebug(sm.Commands.Debug)
	} else if sm.Commands.Version {
		// show version and build
		version.PrintVersion()
	} else if sm.Commands.Verify {
		services := sm.requestedServicesAndProfiles()
		ok := sm.VerifyAllServicesAreRunning(services)
		if !ok {
			os.Exit(13)
		}
	} else if sm.Commands.AutoComplete {
		cli.GenerateAutoCompletions()
	} else {
		// show help if they're not using --update-config with another command
		if !sm.Commands.UpdateConfig {
			fmt.Print(helptext)
		}
	}

	if err != nil {
		fmt.Println(err)
	}

}

// get a list of service names to use in the command.
// profiles are expanded out etc...
func (sm *ServiceManager) requestedServicesAndProfiles() []ServiceAndVersion {

	output := []ServiceAndVersion{}

	for i, s := range sm.Commands.ExtraServices {
		if profileServices, ok := sm.Profiles[s]; ok {
			for _, ps := range profileServices {
				output = append(output, ServiceAndVersion{ps, "", ""})
			}
		} else {
			serviceAndVersion := parseServiceAndVersion(s)
			if i == 0 && sm.Commands.Release != "" {
				serviceAndVersion.version = sm.Commands.Release
			}
			output = append(output, serviceAndVersion)
		}
	}
	return output

}
