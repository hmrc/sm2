package cli

import (
	"flag"
)

type UserOption struct {
	Clean         bool
	Config        string
	Debug         string
	Diagnostic    bool
	ExtraServices []string // other services to start/stop beyond the first one set in --start/--stop
	FromSource    bool
	List          string
	Logs          string
	NoProgress    bool
	Offline       bool
	Port          int
	Ports         bool
	Release       string
	Restart       bool
	ReverseProxy  bool
	Start         bool
	Status        bool
	StatusShort   bool
	StopAll       bool
	Stop          bool
	Verbose       bool
	Version       bool
	Wait          int
	Workers       int
}

func Parse(args []string) *UserOption {

	opts := new(UserOption)

	flagset := flag.NewFlagSet("servicemanager", flag.ContinueOnError)

	flagset.BoolVar(&opts.Clean, "clean", false, "force redownloading of service")
	flagset.StringVar(&opts.Config, "config", "", "use a different config directory")
	flagset.StringVar(&opts.Debug, "debug", "", "infomation on why a given service may not have started.")
	flagset.BoolVar(&opts.Diagnostic, "diagnostic", false, "a suite of checks to debug issues with service manager")
	flagset.BoolVar(&opts.FromSource, "src", false, "run service from source")
	flagset.StringVar(&opts.List, "list", "", "show the content of a profile or searches known services for a match")
	flagset.StringVar(&opts.Logs, "logs", "", "shows the stdout logs for a service")
	flagset.BoolVar(&opts.NoProgress, "no-progress", false, "prevents download progress being shown")
	flagset.BoolVar(&opts.Offline, "offline", false, "starts a service in offline mode.")
	flagset.IntVar(&opts.Port, "port", -1, "overrides the default port for a service")
	flagset.BoolVar(&opts.Ports, "ports", false, "shows which ports services use")
	flagset.StringVar(&opts.Release, "r", "", "sets which version to run")
	flagset.BoolVar(&opts.Restart, "restart", false, "restarts one or more services")
	flagset.BoolVar(&opts.ReverseProxy, "reverse-proxy", false, "starts a reverse proxy to all services on port :3000")
	flagset.BoolVar(&opts.Start, "start", false, "starts one or more service, for a single service use -r to specify version")
	flagset.BoolVar(&opts.Status, "status", false, "shows which services are running")
	flagset.BoolVar(&opts.StatusShort, "s", false, "shows which services are running")
	flagset.BoolVar(&opts.StopAll, "stop-all", false, "stops all services")
	flagset.BoolVar(&opts.Stop, "stop", false, "stops one or more services")
	flagset.BoolVar(&opts.Verbose, "v", false, "enable verbose output")
	flagset.BoolVar(&opts.Version, "version", false, "show the version of service-manager")
	flagset.IntVar(&opts.Wait, "wait", 0, "used with --start, waits a specified number of seconds for the services to become available before exiting")
	flagset.IntVar(&opts.Workers, "workers", 2, "how many services should be downloaded at the same time")
	flagset.Parse(args)

	if opts.Workers <= 0 {
		panic("invalid number of workers set must be > 0")
	}

	// @hack, i didnt want to use a 3rd party arg parser, so we do a sort of hack here of taking the left over args,
	// anything that isnt - prefixed is assumed to be a service, and then if we encounter a - we reparse whats left
	// this allows for sm --start FOO -r 1.2.3 to still work, or sm --start FOO BAZ BAR -v --no-progress
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
		// todo: return a error if a flag is after the service list
	}

	return opts
}
