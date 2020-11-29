#!/bin/sh sh

VERSION=`cat ./VERSION`

echo "Building docker image ..."
docker image build -t dev.volix/thor:${VERSION} .
