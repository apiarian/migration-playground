kafka_dir := $(HOME)/development/kafka
confluent_dir := $(kafka_dir)/latest


zookeeper: ## start the zookeeper for kafka. run this first
	cd $(confluent_dir); ./bin/zookeeper-server-start ./etc/kafka/zookeeper.properties


kafka: ## start kafka. run this after zookeeper is running
	cd $(confluent_dir); ./bin/kafka-server-start ./etc/kafka/server.properties


registry: ## start the schema registry. run this after kafka is running
	cd $(confluent_dir); ./bin/schema-registry-start ./etc/schema-registry/schema-registry.properties


.PHONY: help

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help


# self-documenting makefile:
# http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
