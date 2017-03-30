#!/bin/bash

set -xe

original_topic=`python -c 'import time; print "things-original-{}".format(time.time()),'`
new_topic=`python -c 'import time; print "things-new-{}".format(time.time()),'`

go build -o original-api-api ./original-api/
go build -o shiny-api-api ./shiny-api/

cleanup() {
	killall original-api-api || true
	killall shiny-api-api || true

	rm ./original-api-api
	rm ./shiny-api-api
}

trap cleanup EXIT

./original-api-api -original-topic $original_topic 2>&1 | sed -e 's/^/(original-api) /' &

./shiny-api-api -new-topic $new_topic 2>&1 | sed -e 's/^/(shiny-api) /' &

# sleep forever
while true; do sleep 10000; done
