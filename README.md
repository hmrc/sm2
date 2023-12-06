# Service manager V2
Service Manager 2 (sm2) is a tool to manager starting and stopping groups of scala/play microservices for local development and testing.
It's based on the the original [service-manager](https://github.com/hmrc/service-manager) but re-written in golang.

## Installing Service Manager

### Pre-requisites

`sm2` expects a folder called `service-manager-config` to be present in the workspace (default `$HOME/.sm2/`).

You can find out more [here](example/service-manager-config)

### Installing using `install.sh`

`install.sh` expects an argument that points to the git repository for your `service-manager-config`

Replacing the placeholder git repository with your actual one, run the following in your terminal:

```shell
curl -fksSL https://raw.githubusercontent.com/hmrc/sm2/main/install.sh | bash -s -- "git@github.com:placeholderOrg/placeholderRepo.git"
```

### Installing From Binary

If you'd prefer to carry out these steps manually, then follow these steps:

1. Run the following command in your terminal for your operating system/cpu:

**Linux Intel**
```shell
curl -L -O https://github.com/hmrc/sm2/releases/download/v1.2.0/sm2-1.2.0-linux-intel.zip && unzip sm2-1.2.0-linux-intel.zip && rm sm2-1.2.0-linux-intel.zip && chmod +x sm2
```

**Linux Arm64**
```shell
curl -L -O https://github.com/hmrc/sm2/releases/download/v1.2.0/sm2-1.2.0-linux-arm64.zip && unzip sm2-1.2.0-linux-arm64.zip && rm sm2-1.2.0-linux-arm64.zip && chmod +x sm2
```

**OSX/Apple (latest M1/M2 cpus)**

```shell
curl -L -O https://github.com/hmrc/sm2/releases/download/v1.2.0/sm2-1.2.0-apple-arm64.zip && unzip sm2-1.2.0-apple-arm64.zip && rm sm2-1.2.0-apple-arm64.zip && chmod +x sm2
```

**OSX/Apple (older Intel cpus)**

```shell
curl -L -O https://github.com/hmrc/sm2/releases/download/v1.2.0/sm2-1.2.0-apple-intel.zip && unzip sm2-1.2.0-apple-intel.zip && rm sm2-1.2.0-apple-intel.zip && chmod +x sm2
```

If everything has worked you should have an executable called `sm2`.

2. Move the executable somewhere on your systems $PATH.
3. Run `sm2` from the terminal and follow the instructions to configure your system (if its not already set up)
4. For more information, see [USERGUIDE.md](USERGUIDE.md)

Alternatively you can download and decompress sm2 yourself from the [releases](https://github.com/hmrc/sm2/releases/latest) section. Please note OSX/Apple users will have to [allow `sm2` to run](https://support.apple.com/en-gb/guide/mac-help/mh40616/mac) when downloading via the browser.

### Setup
If you are upgrading from the original service-manager, sm2 will use your existing config.

By default `sm2` will create a workspace folder in `$HOME/.sm2`. If you want to override this default you can do so by following these steps:

1. Create a workspace directory somewhere in your $HOME directory. The directory can be named anything.
```mkdir /home/username/.servicemanager```
2. Set an environment variable called `WORKSPACE` that points to this directory.
```shell
export WORKSPACE=/path/to/workspace
```
Once working add the export to your .bashrc and/or .profile to set it permanently.

2. Using git, clone a copy of service-manager-config into your workspace folder (which will be `$HOME/.sm2` by default unless you've overriden it via `$WORKSPACE`).

### Post install/setup checkers
You can check everything is setup correctly by running:
```shell
sm2 --diagnostic
```

### Enabling tab completion
Run the commands below which will copy a completion script into a file called `sm2.bash` in the directory
`~/.local/share/bash-completion/completions`, if the directory doesn't exist then it will create it.
```shell
mkdir -p ~/.local/share/bash-completion/completions
sm2 --generate-autocomplete > ~/.local/share/bash-completion/completions/sm2.bash
```
If you are using zsh as your shell and not using Oh-My-Zsh then you may need to enable bash-completion support by adding the following to your `~/.zshrc` file
```shell
# Load bash completion functions
autoload -Uz +X compinit && compinit
autoload -Uz +X bashcompinit && bashcompinit
source ~/.local/share/bash-completion/completions/sm2.bash
```


### Upgrading Service Manager 2
As of v1.0.9 `sm2` can update itself - simply run `sm2 -update`. You will need to ensure `sm2` is available on your `$PATH`.

Alternatively, upgrades are a simple matter of downloading the latest version of sm2 and overwriting the `sm2` binary with the new one.

If you are unsure where `sm2` is installed you can use the whereis command to find it:
```shell
$ whereis sm2
sm2 : /usr/local/bin/sm2
```


# Using Service Manager

### Starting a service (-start)
From your terminal type:
```shell
$ sm2 -start SERVICE_NAME
Starting 1 services on 2 workers
 SERVICE_NAME         [====================][100%] Done
```
This will download the latest release of that service, configure it and start it up.
Services names are all uppercase with underscores instead of dashes.

The download will only happen once, after that the service will be cached in your `$WORKSPACE` folder until a new version is released.

### Starting a group of services
Much like starting a single service a group of services (defined by an entry in profiles.json) can be started by typing
```shell
$ sm2 -start PROFILE_NAME
```

Alternatively you can start more than one service at once by typing multiple service and/or profile names
```shell
$ sm2 -start SERVICE_ONE SERVICE_TWO
```

#### Starting a large group of services
Starting a large group of services can overload the cpu of a machine and lead to services failing to start.
If this happens use the following command to start the services at a slower pace.
```shell
$ sm2 --start LARGE_PROFILE_NAME --workers 1 --delay-seconds 5
```
The workers argument starts one service at a time and the DelaySeconds argument adds a 5 second delay inbetween services.

### Starting specific versions
If you need to run a specific version of a service you can do so by adding a colon followed by the version number to the service name, e.g.
```shell
$ sm2 -start SERVICENAME:1.2.3
```

You can also start a specific version of a service using the older `-r` (release) flag:
```shell
$ sm2 -start SERVICENAME -r 1.2.3
```

When starting more than one service, the `-r` flag only applies to the first service in the list.

| Option         | Description                                                                                                          |
|----------------|----------------------------------------------------------------------------------------------------------------------|
| -appendArgs    | A json map of extra args for services being started: `{"SERVICE_NAME":["-DFoo=Bar","SOMETHING"]}`                    |
| -clean         | Deletes the cached version of a service to force a redownload.
| -offline       | Start a service using the cached version. Fails is not in cache. `-offline` can be used by itself to list available services.
| -port 1234     | Overrides the service’s default port to use the supplied port instead.
| -noprogress    | Disabled the progress bars. Useful for scripting and automation.
| -src           | Runs the service(s) from source instead of downloading the binary artifacts. Service manager will attempt to clone the repository and start the service using sbt start. Assumes the system has git configured and a working sbt installation.
| -update-config | Updates workspace copy of service-manager from git. Will fail if there are uncommitted changes or if the config repo is not on the main branch.
| -wait 120      | Waits a given number of seconds for the service to start before exiting.
| -workers 4     | Sets the number of concurrent downloads (default 2). Can also be set via SM_WORKERS environment variable.

### Stopping services (-stop)
```shell
$ sm2 -stop SERVICE_NAME
```

As with the start command you can pass a profile name, or multiple service names to stop more than one service at once.
If you wish to stop all services managed by service manager you can use:
```shell
$ sm2 -stop-all
```

You can also restart services using:
```shell
$ sm2 -restart SERVICE_NAME
```

When restarting a service or profile, you can specify `-latest` to check for a new version before restarting:
```shell
$ sm2 -restart SERVICE_NAME -latest
```


### Checking the status of services (-s or -status)
You can check what services are running using the status command
```shell
$ sm2 -s
+------------------------------------+-----------+---------+-------+--------+
| Name                               | Version   | PID     | Port  | Status |
+------------------------------------+-----------+---------+-------+--------+
| MONGO                              |           | 0       | 27017 |  PASS  |
| SERVICE_FRONTEND                   | 1.73.0    | 64802   | 9057  |  BOOT  |
| SERVICE_CONFIGS                    | 0.130.0   | 24384   | 8460  |  PASS  |
+------------------------------------+-----------+---------+-------+--------+
```

Services will be in one of three states:

| State | Meaning                                                              |
|-------|----------------------------------------------------------------------|
| BOOT  | The service is still starting up and is not ready to use             |
| PASS  | The service has started and its health-check endpoint is responding  |
| FAIL  | The process failed to start, or has started and is no longer running |


Discovering what services are available (-list and -search)
If you are unsure the exact name of a service, want to see what services are available or want to know what services make up a given profile you can use:
```shell
$ sm2 -list
```

to list all available services and profiles, or
```shell
$ sm2 -search SERVICE
```

To search for services and profiles that match the given term. You can supply any valid regex as a parameter to `-search`.

### Discover what port a service uses (-ports)
```shell
sm2 -ports
```
Shows every configured service along with the port it will run on. The output is intended to be easily piped into grep to allow for looking up a specific service or port
```shell
sm2 -ports | grep CATALOGUE_FRONTEND
# or
sm2 -ports | grep 9540
```

## Troubleshooting Service Manager
Sometimes a service will fail to start up. To help determine why, service manager has some built-in features to help diagnose failing services.

### Diagnostic Mode (-diagnostic)
Before doing anything else, it’s worth running service-manager’s self-checks to ensure it is installed correctly.
```shell
$ sm2 -diagnostic
version: 1.2.0
  build: ef49b60
OS:		 OK (linux, amd64)
JAVA:		 OK (11.0.17)
GIT:		 OK (git version 2.38.1)
NET		 OK (VPN check timeout 20s)
VPN DNS		 OK (IP Address of artefactory resolvable
VPN:		 OK (artifactory/api/system/ping responds to ping)
WORKSPACE:	 OK (/home/user.servicemanager)
CONFIG:		 WARN: Local version (abe9d50) is not up to date with remote version (eaaa410)
```

This will do a number of checks to ensure sm2 is able to download and install services.
Also when raising a support request it is often helpful to include the output of this command.

### Debug Mode (-debug SERVICE_NAME)
If your service is failing to start, you can get a detailed breakdown of what was attempted and what succeeded using the -debug flag.
```shell
$ sm2 -debug SERVICE_NAME

Checking .install file...
SERVICE_CONFIGS: version 0.130.0
 Installed at /home/user/.servicemanager/install/service-configs/service-configs-0.130.0 on 2022-11-11 10:31:21.310173617 +0000 UTC
Checking .state file...
The .state file says SERVICE_CONFIGS version 0.130.0 was started on 2022-11-11 10:31:21.310730227 +0000 UTC with PID 24384
It was run with the following args:
	- -Dconfig.resource=application.conf
	- -Dapplication.router=testOnlyDoNotUseInAppConf.Routes
	- -J-Xmx256m
	- -J-Xms64m
	- -Dservice.manager.serviceName=SERVICE_CONFIGS
	- -Dservice.manager.runFrom=0.130.0
	- -Duser.home=/home/user/.servicemanager/install/service-configs
	- -Dhttp.port=8460
Checking pid: 24384 is running...
Pid 24384 exists, service is probably running...pinging service on port 8460...
Service responded to ping on [http://localhost:8460/ping/ping], its alive.
Log files in /home/user/.servicemanager/hmrc/install/service-configs/service-configs-0.130.0/logs:
	          access.log  0
	       connector.log  0
	 service-configs.log  7099
	          stdout.log  8415

```

Debug mode checks what was requested, what was actually installed, what was started and with what parameters and if there are any logs or healthcheck responses.

### Viewing Logs (-logs SERVICE_NAME)
If a service is running and you simply want to check the stdout/stderr of the process you can view it using:
```shell
sm2 -logs SERVICE_NAME
```


## Developing using service-manager

Ok so you’ve installed service manager and started some services, fantastic, but how does this fit into your development process?

Generally we’d recommend the following
1. Checkout the source code for the service you’re working on
2. Start the service from the IDE
3. Use service manager to start the profile related to the service your developing
4. Service manager should skip starting the service running in the IDE as its healthcheck should respond telling SM its already running.

Alternatively you can start the profile beforehand and just stop the individual service you want to work on.

## ASSETS_FRONTEND
Assets Frontend is a service that only exists in service-manager responsible for serving up static assets (that would normally come from a CDN) to local services.

The original assets-frontend (as defined in services.json as ASSETS_FRONTEND) will not work in service manager 2. However, a new service, ASSETS_FRONTEND_2 is available as a drop-in replacement offering

### New version
Assets frontend 2 is a drop-in replacement for Assets Frontend.

The new version of assets frontend ASSETS_FRONTEND_2 is a complete rewrite. It is a scala service, so no longer requires a specific version of python to work and will only download the assets when they are requested rather than downloading 100s of megabytes of assets on start-up or being limited to one specific version.

### How to start ASSETS_FRONTEND_2
It can be started the same as any other service.
```
$ sm2 -start ASSETS_FRONTEND_2
```
You do not need to specify which version of the asset you wish to use, just run your service(s) as normal and ASSETS_FRONTEND_2 will download them when requested.
The individual asset bundles are little more than 1 meg in size and will be cached so the download time should not be noticeable.

Assets frontend 2 will also work in the original service manager, so there's no reason not to update your profiles to use it.

## Configuration
### Environment Variables
Service manager has a few options that can be set via environment variables. Typically you should set these by adding an export statement to your .bash_profile or .profile file in your home directory (this will vary depending on your operating system)
```shell
export WORKSPACE=/home/myusername/.servicemanager
```


| Environment Variable | Description                                                                                           |
|----------------------|-------------------------------------------------------------------------------------------------------|
| WORKSPACE | (required) Path to service managers workspace folder. Cached services and config will be stored here  |
| SM_TIMEOUT | Overrides the default http timeouts. Useful if you have a very slow internet connection |
| SM_WORKERS | Sets the number of concurrent downloads. Same as using the -workers flag. |

 ### Service Manager Config
To run service manager you will require a folder named service-manager-config to exist inside your WORKSPACE folder. It should typically be a clone of a git repository.
Service-manager-config is expected to have the following structure:

| File          | Description                                                |
|---------------|------------------------------------------------------------|
| config.json   | Defines the urls for artifactory                           |
| services.json | Defines all the available services                         |
| profiles.json | Defines groups of services that should be started together |

### Setting Scala Version
For Scala artifacts, the artifact name will include the Scala version:

```json
    "binary": {
      "artifact": "example-service_2.12"
    }
```

However, you may represent the Scala version with `_%%` instead. Then service-manager will use the latest artifact version it finds, regardless of Scala version. This should make it simpler to maintain:


```json
    "binary": {
      "artifact": "example-service_%%"
	}
```


## Building/Developing Service-Manager-2
SM2 has no external dependencies other than go 1.20+. You can build it locally via:

```bash
go build
```

When building a new release, use the included makefile:
```bash
make build_all package
```

## Running tests

To run all tests in all sub-directories, ensure you are in the project root and run:

```bash
make test
```

To run tests for the subdirectory you are currently in:
```bash
go test
```

## Build with Nix

To install from source:

```bash
nix-env -i -f default.nix
```

## For maintainers

nixpkgs is pinned. To update - edit branch in `update-nixpkgs.sh`, then regenerate `nixpkgs.json`:

```bash
bash update-nixpkgs.sh
```

To build with shell:
```bash
nix-shell --run 'go build'
```
to create `./sm2`

or to build locally:

```bash
nix-build
```
to create `./result/bin/sm2`
