#!/bin/bash

set -e

hackdir=$(CDPATH="" cd $(dirname $0); pwd)

# If we are running inside of Travis then do not run the rest of this
# script unless we want to TEST_ASSETS
if [[ "${TRAVIS}" == "true" && "${TEST_ASSETS}" == "false" ]]; then
  exit
fi

pushd ${hackdir}/../assets > /dev/null
  bundle exec grunt test
  bundle exec grunt build
popd > /dev/null

pushd ${hackdir}/../assets > /dev/null
  echo ""
  echo "Source asset checksums..."
  find .tmp -type f | sort | xargs md5sum

  echo ""
  echo "Built asset checksums..."
  find dist -type f | sort | xargs md5sum
popd > /dev/null  

pushd ${hackdir}/../ > /dev/null
  Godeps/_workspace/bin/go-bindata -prefix "assets/dist" -pkg "assets" -o "_output/test/assets/bindata.go" assets/dist/...
  echo "Validating checked in bindata.go is up to date..."
  if ! diff -q _output/test/assets/bindata.go pkg/assets/bindata.go ; then

    pushd ${hackdir}/../assets > /dev/null

      if [ -f debug.zip ]; then
        unzip debug.zip -d debug
        diff -r .tmp debug/.tmp
        diff -r dist debug/dist
      fi

      if [[ "${TRAVIS}" == "true" ]]; then
        echo ""
        echo "Bundler versions..."
        bundle list

        echo ""
        echo "Bower versions..."
        bower list -o

        echo ""
        echo "NPM versions..."
        npm list
      fi

    popd > /dev/null  

    exit 1
  fi
popd > /dev/null