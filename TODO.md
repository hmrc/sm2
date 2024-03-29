=TODO

===nice to have
- windows support (just needs platform impl of uptime and pids)
- better error reporting on service startup failure
- override JAVA_HOME based on config/known  JRE locations
- template generator for service/profiles
- option to load default mongo data

===todo
- check service type on startup, better error for non-play
- server mode
- user level config (i.e. override tmpdir, default worker count, artifactory url, vpn check etc)
- git pull when running from src (use .install)
- seperate output from await and stop actions
- return error code on --status if any are not working


=== done
- vpn check, use ping endpoint
- integration tests
- make / configurable for reverse-proxy (i.e swap between catalogue etc)
- deduped assets frontend || on-demand assets frontend
- command to update config
- support start multiple services with versions e.g. --start SERVICE_FOO:1.4.0 FIZZBUZZ:0.555.0
- checkports command
- validate md5sum on download (artifactory hash)
- use service.status to restart services
- support --appendArgs and its weird json encoded payload
- mocks for testing artifactory
- support the -wait parameter
- offline mode (may need refactoring? or different start path, would need to scrap .install and use that instead of artifactory)
- debug mode
- dedupe service list when starting multiple services
- check .state as part of --status
- could artifactory un-tar track which dirs its created and use that to find the service-dir? (test)
- startfromsource to show progress correctly
- validate service.state when deciding weather or not to re-download a service (i.e. handle corrupt downloads)
- throttle multi-downloads to n at a time...
- first pass at downloader workgroups
- download more than one service at a time
- improve progress tracking
- show progress as a percentage
- make cli start/stop bools and use unparsed args as service list (?)
- move profile expansion into commands, do away with Start(profileOrService) in favor of just Start
- support starting more than one service per command
- asset frontend server
- run from source (rebase/update)
- remove old versions
- add --restart command
- improve status command, get version etc
- fix user path so service logs arent written to the working dir (test)
- -port flag support
- no-progress flag
- clean running.pid if a service is downloaded but not running
- mongo status check (todo: perf)
- dont download if already downloaded
- show the logs for a service
- fix log location
- start a profile
- stop a profile
- add extra args when starting a service (port, sm id user supplied)
- start latest cmd
- run version -r
- cli interface
- parse artifactory maven-metadata.xml
- stop a service by name
- start a service after downloading it
- load sm config
- show running services
- show all ports
- download a service by name
- normal check checks
- vpn check
- reverse proxy mode
- diagnostic flag to check vpn/git/path/config etc
- improve diagnostic checks, add a flag
- print version number
