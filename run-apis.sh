#!/bin/bash

set -xe

command_topic=`python -c 'import time; print "thing-commands-{}".format(time.time())'`
original_topic=`python -c 'import time; print "things-{}".format(time.time())'`

go build -o original-api-api ./original-api/
go build -o shiny-api-api ./shiny-api/api/
go build -o shiny-api-updater ./shiny-api/updater/

cleanup() {
	killall original-api-api || true
	killall shiny-api-api || true
	killall shiny-api-updater || true

	rm ./original-api-api
	rm ./shiny-api-api
	rm ./shiny-api-updater
}

trap cleanup EXIT

./original-api-api -original-topic $original_topic 2>&1 | sed -e 's/^/(original-api) /' &

./shiny-api-api -command-topic $command_topic 2>&1 | sed -e 's/^/(shiny-api) /' &

./shiny-api-updater -command-topic $command_topic 2>&1 | sed -e 's/^/(updater) /' &

# sleep forever
while true; do sleep 10000; done
