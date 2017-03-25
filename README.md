# Migration Playground

A place to run experiments on migrating from one API backend to another using
Kafka topics to choreograph the multi-master dance.

## Local Kafka

### Setup

1. Download the [Confluent Open Source Platform](https://www.confluent.io).
1. Extract the tar or zip file in the `~/development/kafka` directory
1. Make a symlink to the latest confluent directory at `~/development/kafka/latest`

### Running Kafka

1. Open three terminal windows at the `migration-playground` directory
1. Run `make zookeeper` in the first and wait for zookeeper load
1. Run `make kafka` in the second and wait for kafka to load
1. Run `make registry` in the third and wait for the schema registry to load
1. When you're ready to shut things down, send an Interrupt (`^C`) to each of
the terminal windows in reverse order

