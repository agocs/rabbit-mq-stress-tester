#!/usr/bin/env bash

VERSION=`git rev-parse --abbrev-ref HEAD`-`git rev-parse --short HEAD`
VERSION='1.2.3'

echo "Building docker version... $VERSION"

docker build -t docker-registry.management.fravega.com/fravega/rabbit-mq-stress-tester:${VERSION} . \
|| { echo 'Building docker image and tags failed' ; exit 1; }

docker push docker-registry.management.fravega.com/fravega/rabbit-mq-stress-tester:${VERSION} \
|| { echo 'Pushing image to server failed' ; exit 1; }
