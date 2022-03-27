package servicemanager

import (
	"fmt"
	"io/ioutil"
	"path"
	"regexp"
	"sort"
	"strings"

	"sm2/ledger"
)

type portListing struct {
	port     int
	service  string
	frontend bool
}

func (sm *ServiceManager) ListPorts() {
	output := []portListing{}

	maxLen := 20
	for _, v := range sm.Services {
		if len(v.Id) > maxLen {
			maxLen = len(v.Id)
		}
		output = append(output, portListing{v.DefaultPort, v.Id, v.Frontend})
	}

	sort.Slice(output, func(i, j int) bool {
		return output[i].port < output[j].port
	})

	frontend := ""
	for _, o := range output {
		if o.frontend {
			frontend = "*"
		}
		fmt.Printf("%-5d -> %s  %s\n", o.port, pad(o.service, maxLen), frontend)
	}
}

func (sm *ServiceManager) ListServices(filter string) {

	// check if its a profile, list services and exit
	if profile, ok := sm.Profiles[strings.ToUpper(filter)]; ok {
		fmt.Printf("Profile %s has these services:\n", strings.ToUpper(filter))
		for _, p := range profile {
			fmt.Printf("%s\n", p)
		}
		return
	}

	// check if its an exact match to a service
	if service, ok := sm.Services[strings.ToUpper(filter)]; ok {
		fmt.Println("Found exact match for service:")
		fmt.Printf("%-25s -> %s\n\n", service.Id, service.Name)
	}

	// else search the services for likely matches

	// extract and sort keys
	keys := make([]string, len(sm.Services))
	longestKey := 0
	i := 0
	for k := range sm.Services {
		keys[i] = k
		if len(k) > longestKey {
			longestKey = len(k)
		}
		i++
	}
	sort.Strings(keys)

	// build regex
	search := regexp.MustCompile(fmt.Sprintf(".*%s.*", strings.ToUpper(filter)))

	// run search
	fmt.Printf("Searching for (%s)...\n", search.String())
	for _, k := range keys {
		if search.MatchString(k) {
			if service, ok := sm.Services[k]; ok {
				fmt.Printf("%s -> %s\n", pad(service.Id, longestKey), service.Name)
			}
		}
	}
}

// scrapes the install files and prints out what versions are installed and available
func (sm *ServiceManager) ListServicesAvailableOffline() {

	files, err := ioutil.ReadDir(sm.Config.TmpDir)

	if err != nil {
		fmt.Printf("failed to read the workspace dir, %s\n", err)
		return
	}

	matches := []ledger.InstallFile{}
	for _, file := range files {
		if file.IsDir() {
			if state, err := sm.Ledger.LoadInstallFile(path.Join(sm.Config.TmpDir, file.Name())); err == nil {
				matches = append(matches, state)
			}
		}
	}

	if len(matches) == 0 {
		fmt.Println("No services are installed or are available offline.")
		return
	}

	fmt.Println("The following services are installed and available offline:")
	for _, install := range matches {
		fmt.Printf(" %-25s %-9s installed\n", install.Service, install.Version)
	}
	fmt.Println("You can start them using sm2 --offline --start SERVICE_NAME")

}
