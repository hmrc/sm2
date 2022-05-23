package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"strings"
)

type UserOption struct {
	appendArgs    string              // not exported, content decoded into ExtraArgs
	CheckPorts    bool                // finds duplicate ports
	Clean         bool                // used with --start to force redownloading
	Config        string              // uses a different service-manager-config folder
	Debug         string              // debug info about a service, used to determin why it failed to start
	Diagnostic    bool                // runs tests to determin if there are problems with the install
	ExtraArgs     map[string][]string // parsed from content of AppendArgs
	ExtraServices []string            // ids of services to start
	FromSource    bool                // used with --start to run from source rather than bin
	List          bool                // lists all the services
	Logs          string              // prints the logs of a service, running or otherwise
	NoProgress    bool                // hides the animated download progress meter
	Offline       bool                // prints downloaded services, used with --start bypasses download and uses local copy
	Port          int                 // overrides service port, only works with the first service when starting multiple
	Ports         bool                // prints all the ports
	Release       string              // specify a version when starting one service. unlikely old sm, cannot be used without a version
	Restart       bool                // restarts a service or profile
	ReverseProxy  bool                // starts a reverse-proxy on 3000 (override with --port)
	Search        string              // searches for services/profiles
	Start         bool                // starts a service, multiple services or a profile(s)
	Status        bool                // shows status of everything thats running
	StatusShort   bool                // same as --status but is the -s short version of the cmd
	StopAll       bool                // stops all the services that are running
	Stop          bool                // stops a service, multiple services or profile(s)
	Verbose       bool                // shows extra logging
	Version       bool                // prints sm2 version number
	Wait          int                 // waits given number of secs after starting services for then to respond to pings
	Workers       int                 // sets the number of concurrent downloads/service starts
}

func Parse(args []string) (*UserOption, error) {

	opts := new(UserOption)

	flagset := flag.NewFlagSet("servicemanager", flag.ExitOnError)
	flagset.StringVar(&opts.appendArgs, "appendArgs", "", "A map of args to append for services you are starting. i.e. '{\"SERVICE_NAME\":[\"-DFoo=Bar\",\"SOMETHING\"],\"SERVICE_TWO\":[\"APPEND_THIS\"]}'")
	flagset.BoolVar(&opts.CheckPorts, "checkports", false, "finds services using the same port number")
	flagset.BoolVar(&opts.Clean, "clean", false, "forces reinstall of service (use with --start)")
	flagset.StringVar(&opts.Config, "config", "", "sets an alternate directory for service-manager-config")
	flagset.StringVar(&opts.Debug, "debug", "", "infomation on why a given `service` may not have started")
	flagset.BoolVar(&opts.Diagnostic, "diagnostic", false, "a suite of checks to debug issues with service manager")
	flagset.BoolVar(&opts.FromSource, "src", false, "run service from source (use with --start)")
	flagset.BoolVar(&opts.List, "list", false, "lists all available services")
	flagset.StringVar(&opts.Logs, "logs", "", "shows the stdout logs for a service")
	flagset.BoolVar(&opts.NoProgress, "noprogress", false, "prevents download progress being shown (use with --start)")
	flagset.BoolVar(&opts.Offline, "offline", false, "starts a service in offline mode (use with --start or standalone to list available services)")
	flagset.IntVar(&opts.Port, "port", -1, "overrides the default port for a service (use with --start)")
	flagset.BoolVar(&opts.Ports, "ports", false, "shows which ports services use")
	flagset.StringVar(&opts.Release, "r", "", "sets which `version` to run (use with --start)")
	flagset.BoolVar(&opts.Restart, "restart", false, "restarts one or more services")
	flagset.BoolVar(&opts.ReverseProxy, "reverse-proxy", false, "starts a reverse proxy to all services on port :3000")
	flagset.StringVar(&opts.Search, "search", "", "searches for services and profiles that match a given `regex`")
	flagset.BoolVar(&opts.Start, "start", false, "starts one or more service, for a single service use -r to specify version")
	flagset.BoolVar(&opts.Status, "status", false, "shows which services are running")
	flagset.BoolVar(&opts.StatusShort, "s", false, "shows which services are running")
	flagset.BoolVar(&opts.StopAll, "stop-all", false, "stops all services")
	flagset.BoolVar(&opts.Stop, "stop", false, "stops one or more services")
	flagset.BoolVar(&opts.Verbose, "v", false, "enable verbose output")
	flagset.BoolVar(&opts.Version, "version", false, "show the version of service-manager")
	flagset.IntVar(&opts.Wait, "wait", 0, "used with --start, waits a specified number of seconds for the services to become available before exiting (use with --start)")
	flagset.IntVar(&opts.Workers, "workers", 2, "how many services should be downloaded at the same time (use with --start)")
	flagset.Parse(args)

	if opts.Workers <= 0 {
		return nil, fmt.Errorf("invalid number of workers set must be > 0")
	}

	// @hack, i didnt want to use a 3rd party arg parser, so we do a sort of hack here of taking the left over args,
	// anything that isnt - prefixed is assumed to be a service, and then if we encounter a - we reparse whats left
	// this allows for sm --start FOO -r 1.2.3 to still work, or sm --start FOO BAZ BAR -v --noprogress
	serviceSeen := map[string]bool{}

	for i, arg := range flagset.Args() {
		if len(arg) > 0 && arg[0] != '-' {
			if _, seen := serviceSeen[arg]; !seen {
				opts.ExtraServices = append(opts.ExtraServices, arg)
				serviceSeen[arg] = true
			}
		} else {
			flagset.Parse(flagset.Args()[i:])
			break
		}
	}

	// Decode appendArgs (to keep legacy compatibility they're encoded as json for some reason)
	if opts.appendArgs != "" {
		args, err := parseAppendArgs(opts.appendArgs)
		if err != nil {
			return nil, fmt.Errorf("problem decoding --appendArgs: %s, check --help for format", err)
		}
		opts.ExtraArgs = args
	}
	return opts, nil
}

/*
   Parses extra args for all the services. Expected format is:
   {"SERVICE_NAME":["-DFoo=Bar","SOMETHING"],"SERVICE_TWO":["APPEND_THIS"]}
*/
func parseAppendArgs(jsonArgs string) (map[string][]string, error) {

	args := map[string][]string{}

	decoder := json.NewDecoder(strings.NewReader(jsonArgs))
	err := decoder.Decode(&args)

	return args, err
}
