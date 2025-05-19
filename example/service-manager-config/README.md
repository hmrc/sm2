# Example Config

Service Manager expects a folder called `service-manager-config` to be present in its workspace folder (`~/.sm2`).
The config folder should be structured like this:

```
- config.json
- services.json
- services/*.json (optional)
- profiles.json

```

### config.json
This file defines how to connect to artifactory. Should point to where ever the artifacts are being hosted.

### services.json
A json map describing all the services that can be run by service-manager. 
The key for each map entry is the ID service-manager will use to manage the service.
The source section is not required and can be omitted if you dont need to run from source.

### multiple json files in a services folder
In the case that your services.json becomes hard to maintain due to the number of services defined. Service definitions can be split across multiple files and placed in a folder called `services`

Presence of the `services` folder means that `services.json` will be ignored.

### profiles.json
A json map describing groups of services that can be started using a single command. The key will be the profile name and the values will be an array of service names (defined in services.json). 
