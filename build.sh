#!/bin/sh sh
# Simple script to add version and build information
# to a Go application.
VERSION=`cat ./VERSION`
DATE=`date +%FT%T%z`
HOST=`hostname`
USER=`git config user.name`
GIT_BRANCH=`git symbolic-ref -q --short HEAD`
GIT_COMMIT=`git show -s --format=%h HEAD`
GIT_COMMIT_TIME=`git show -s --format=%ci HEAD`
GIT_SUMMARY=`git describe --tags --dirty --always`

echo "Building go application version ${VERSION} @${GIT_COMMIT} ..."
start=`date +%s%3N`

# Your full package path here.
PKG="dev.volix.ops/thor/pkg/version"

go build -o thor -ldflags "-X '${PKG}.Version=${VERSION}' \
  -X '${PKG}.BuildTime=${DATE}' \
  -X '${PKG}.BuildHost=${HOST}' \
  -X '${PKG}.BuildUser=${USER}' \
  -X '${PKG}.GitBranch=${GIT_BRANCH}' \
  -X '${PKG}.GitCommit=${GIT_COMMIT}' \
  -X '${PKG}.GitCommitTime=${GIT_COMMIT_TIME}' \
  -X '${PKG}.GitSummary=${GIT_SUMMARY}'" .

if [ $? -eq 0 ]
then
  end=`date +%s%3N`
  echo "Done. (took `expr $end - $start`ms)"
  exit 0
else
  echo "Error during Go building process."
  exit 1
fi
