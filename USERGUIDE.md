# Service Manager v2

Service Manager is a command line tool to start and stop microservices locally.
It is intended to make local development and testing easier.

## Install

Download the latest version for your operating system and architecture.
For ease of use, rename the binary to sm (or sm2 if you wish to keep the original service manager) and put it somewhere
int your systems $PATH.
Finally `chmod +x` the file to make it runnable.

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

| Option          | Description                                                                                                          |
|-----------------|----------------------------------------------------------------------------------------------------------------------|
| `-r 1.0.0`      | Starts a specific release of a service. When starting multiple services the flag only applies to the first service.  |
| `--src`         | Start a service from source. Requires git and sbt to be installed.                                                   |
| `--port 1234`   | Overrides the default port of the service.                                                                           |
| `--no-progress` | Surpresses the progress bars when downloading the service. Suitable for scripts etc.                                 |
| `--offline`     | Starts services that are already without attempting to download the latest version                                   |
| `--wait 20`     | Waits a specified number of seconds for all services to reach a healthy state                                        |

## Stopping a Service

A running service can be stopped with the --stop command:

```
sm2 --stop SERVICE_NAME
```

All running services can be stopped at the same time using:
```
sm2 --stop-all
```

==Seeing the status of running services

The `--status` command (`-s` for short) shows the status of all services that are running or should be running.

```
example table
```

A more details breakdown of the state of a given service can be found using:
```
sm2 --debug SERVICE_NAME

sm2 --logs SERVICE_NAME
```
This can be useful in determining why a service failed to start.

## Listing Services
To discover which services are available to run you can use the `--list` command.
You can discover what services will be run as part of a service profile with `--list PROFILE_NAME`.
If you are unsure of the exact name of a service you can search for likely matches using `--list FOO`, which will show all services containing 'FOO'.
A full list of services can be found using `--list .`

The ports command `--ports` will list all of the services and their default ports.
If you need to run service manager without internet connectivity, running the `--offline` command by itself will list which services are currently installed and avilable for offline use.
Services can be started in offline mode using `--start SERVICE_NAME --offline`.

## Reverse Proxy

A new feature in version 2 is the reverse proxy mode. Using the --reverse-proxy option starts an http server running on port 3000.
Any service that has a valid `location` entry in services.json will be available on port 3000 under that path.
This can be useful if a frontend service needs to pass cookies etc to another frontend service. Often browsers will prevent cookies being passed between hosts and can consider different port on the same host as being distinct hosts.

## Diagnostic Mode
Running `sm2 --diagnotic` will perform some basic health checks for the sm2 tool. It can help diagnose connectivity and configuration issues.

