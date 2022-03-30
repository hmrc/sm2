package servicemanager

const helptext string = `
== Service Manager 2 ==

Start a service:
    sm2 --start AUTH

Start more than one service
    sm2 --start CATALOGUE_FRONTEND INTERNAL_AUTH

Start a specific version of a service:
    sm2 --start AUTH -r 4.0.0

Force a service to be redownloaded:
   sm2 --start AUTH --clean

Show all running services:
    sm2 -s

If you're having trouble with service manager or a service isnt starting, try these commands:
    sm2 --diagnostic
    sm2 --debug SERVICE_NAME
    sm2 --logs SERVICE_NAME

To see a list of all the commands:
    sm2 --help
`
