#!/bin/#!/usr/bin/env bash

docker run -it --rm --name tester docker-registry.cluster.fravega.com/fravega/rabbit-mq-stress-tester:1.1.2 ./rabbit-mq-stress-tester --help
