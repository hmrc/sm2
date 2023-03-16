# Service Manager v2

Service Manager is a command line tool to start and stop microservices locally.
It is intended to make local development and testing easier.

## Setup

1. Create the workspace
Create a workspace directory. Service manager will use this for storing its copy of service-manager-config as well as the services it will run.
Once created, set the WORKSPACE environment variable to point to this folder:
```
export WORKSPACE=/path/to/workspace
```
Once working you should add this to your .bashrc and/or .profile so that it is always set.

2. Get the config
Service manager expects a copy of service-manager-config to be present in the WORKSPACE folder.
Install it via git:
```
git clone git@github.com:hmrc/service-manager-config.git $WORKSPACE/service-manager-config
```

## Starting Services

```
sm2 --start SERVICE_NAME
```

SERVICE_NAME must exist in the services.json file in service-manager-config. If valid service manager will download
the latest version of the service from artifactory and attempt to start it.

Multiple services can be started in one go by passing in more than one service name.
```
sm2 --start SERVICE_ONE SERVICE_TWO SERVICE_THREE
```

| Option            | Description                                                                                                          |
|-------------------|----------------------------------------------------------------------------------------------------------------------|
| `-r 1.0.0`        | Starts a specific release of a service. When starting multiple services the flag only applies to the first service.  |
| `--src`           | Start a service from source. Requires git and sbt to be installed.                                                   |
| `--port 1234`     | Overrides the default port of the service.                                                                           |
| `--noprogress`    | Supresses the progress bars when downloading the service. Suitable for scripts etc.                                  |
| `--offline`       | Starts services that are already without attempting to download the latest version                                   |
| `--clean`         | Removes existing install, forcing a re-download                                                                      |
| `--wait 20`       | Waits a specified number of seconds for all services to reach a healthy state                                        |
| `--appendArgs`    | A json map of extra args for services being started: `{"SERVICE_NAME":["-DFoo=Bar","SOMETHING"]}`                    |
| `--workers 4`     | The number of services to download/start at the same time (default 2)                                                |
| `--reverse-proxy` | Starts a reverse proxy                                                                                               |


Alternatively, instead of setting the version with the `-r` flag, you can start a specific release using the following syntax:
```
sm2 --start SERVICE_ONE:1.2.0 SERVICE_TWO:0.40.0
```

If the version you want to start was build in an older version to the one defined in services.json, you can tell sm2 which version to use by append the scala version to the end of the artifact name:

```
sm2 --start SERVICE_ONE_2.11:0.44.0
```

## Stopping a Service

A running service can be stopped with the --stop command:

```
sm2 --stop SERVICE_NAME
```

All running services can be stopped at the same time using:
```
sm2 --stop-all
```

## Seeing the status of running services

The `--status` command (`-s` for short) shows the status of all services that are running or should be running.

```
+------------------------------------+-----------+---------+-------+--------+
| Name                               | Version   | PID     | Port  | Status |
+------------------------------------+-----------+---------+-------+--------+
| MONGO                              |           | 0       | 27017 |  PASS  |
| INTERNAL_AUTH                      | 0.95.0    | 589265  | 8470  |  FAIL  |
| SAVE4LATER                         | 1.39.0    | 589264  | 9272  |  BOOT  |
| SERVICE_CONFIGS                    | 0.115.0   | 588896  | 8460  |  PASS  |
+------------------------------------+-----------+---------+-------+--------+
```

## Debugging a failed service
A more details breakdown of the state of a given service can be found using:
```
sm2 --debug SERVICE_NAME

sm2 --logs SERVICE_NAME
```
This can be useful in determining why a service failed to start.

## Listing Services
To discover which services are available to run you can use the `--search` command.
You can discover what services will be run as part of a service profile with `--search PROFILE_NAME`.
If you are unsure of the exact name of a service you can search for likely matches using `--search FOO`, which will show all services containing 'FOO'. You can also use regex expressions.
A full list of services can be found using `--search .` or just `--list`.

The ports command `--ports` will list all of the services and their default ports.
If you need to run service manager without internet connectivity, running the `--offline` command by itself will list which services are currently installed and avilable for offline use.
Services can be started in offline mode using `--start SERVICE_NAME --offline`.

## Reverse Proxy
A new feature in version 2 is the reverse proxy mode. Using the --reverse-proxy option starts an http server running on port 3000.
Any service that has a valid `proxyPaths` entry in services.json will be available on port 3000 under that path. 
The `proxyPaths` setting consists of an array of strings, each string representing a path to be proxied to that service.
This can be useful if a frontend service needs to pass cookies etc. to another frontend service.
Often browsers will prevent cookies being passed between hosts and can consider different port on the same host as being distinct hosts.

By default, the reverse-proxy will forward any proxy-path it finds. e.g.
```
sm2 --reverse-proxy
```

You may restrict this behaviour to only a few services or profiles by listing them. e.g.
```
sm2 --reverse-proxy CATALOGUE
```

## Diagnostic Mode
Running `sm2 --diagnostic` will perform some basic health checks for the sm2 tool. It can help diagnose connectivity and configuration issues.

## Keeping service-manager-config up to date
You can use service manager to get the latest config using the `sm2 --update-config` command. It requires the copy of service-manager-config in your $WORKSPACE be on the HEAD branch, if it is not it will not perform the update (so as not to overwrite any changes you may be working on etc).

## Config Options
You can override some of the default settings in sm2 using environment variables.
These environment variables can either be set temporarily in your shell, or added to .profile or .bashrc etc to apply them permanently.

### Number of concurrent downloads
The number of services that will be installed at the same time (default 2) can be set via the `SM_WORKERS` environment variable, e.g.

```
export SM_WORKERS=4
```

would download and install 4 services at the same time. It can be worth tweaking this setting depending on your internet connection.
This is the same as using the --workers argument.

### Increasing the default timeout
The timeout (in seconds) for the vpn check and other calls can be changed via the `SM_TIMEOUT` environment variable, e.g.

```
export SM_TIMEOUT=30
```

would set the timeout to 30 seconds. The vpn check is performed every time you attempt to start a service, unless the `--offline` flag is used.

### Disabling the vpn check
The vpn check can be disabled completely if it is causing issues or for testing via `SM_NOVPN`, e.g.

```
export SM_NOPVN=true
```

This would completely disable the vpn connectivity check when installing a service. The install will still fail if the VPN is not connected.
This is the same as using the --no-vpn-check argument.
