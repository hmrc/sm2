package servicemanager

import (
	"fmt"
	"sync"
	"time"

	"sm2/version"
)

type ServiceAndVersion struct {
	service    string
	version    string
	fromSource bool
}

func (sm *ServiceManager) Run() {

	var err error

	if sm.Commands.Status || sm.Commands.StatusShort {
		// prints table of running services
		sm.PrintStatus()
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
		sm.ListServices(sm.Commands.Search)
	} else if sm.Commands.List {
		// alias for search everything
		sm.ListServices(".")
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
	} else {
		// show help
		fmt.Print(helptext)
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
				output = append(output, ServiceAndVersion{ps, "", sm.Commands.FromSource})
			}
		} else {
			version := ""
			if i == 0 {
				version = sm.Commands.Release
			}
			output = append(output, ServiceAndVersion{s, version, sm.Commands.FromSource})
		}
	}
	return output

}

func (sm *ServiceManager) startServiceWorker(tasks chan ServiceAndVersion, wg *sync.WaitGroup) {

	for task := range tasks {

		var err error
		if task.fromSource {
			err = sm.StartFromSource(task.service)
		} else {
			err = sm.StartService(task.service, task.version)
		}

		if err != nil {
			sm.progress.update(task.service, 100, err.Error())
		} else {
			sm.progress.update(task.service, 100, "Done")
		}
		wg.Done()
	}

}

// Starts a bunch of services at once, but not all at once...
// the serviceWorkers run in concurrently, starting services as they arrive on the
// channel. The renderer also runs concurrently, drawing input as it gets it.
// A wait group is used to keep the app waiting for everything to finish downloading.
func (sm *ServiceManager) asyncStart(services []ServiceAndVersion) {

	// fire up the progress bar renderer
	sm.progress.noProgress = sm.Commands.NoProgress
	go sm.progress.renderLoop()
	sm.progress.init(services)
	taskQueue := make(chan ServiceAndVersion, len(services))

	fmt.Printf("Starting %d services on %d workers\n", len(services), sm.Commands.Workers)

	// start up a number of workers (controlled by --workers param)
	wg := sync.WaitGroup{}
	for i := 0; i < sm.Commands.Workers; i++ {
		go sm.startServiceWorker(taskQueue, &wg)
	}

	for _, sv := range services {
		wg.Add(1)
		taskQueue <- sv
	}

	wg.Wait()
	// @hack @hack waits a ms in the hope the renderloop finishes.
	// this could be way better, wait groups, or force a final paint or something??
	time.Sleep(time.Millisecond)

	if sm.Commands.Wait > 0 {
		fmt.Printf("Waiting %d secs for all services to start.", sm.Commands.Wait)
		sm.Await(services, sm.Commands.Wait)
	} else {
		fmt.Println("Done")
	}
}
