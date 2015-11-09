# helpers to run bats

__FILE__="${BASH_SOURCE[0]}"
GAOL_DIR=$( cd "$(dirname "${BASH_SOURCE[0]}")"/.. && pwd )

if [ -z "$GAOL_TARGET" ]; then
  echo 'please set $GAOL_TARGET to an empty garden server'
  exit 1
fi

load $GAOL_DIR/test/support/assertions.bash

teardown() {
  gaol list | xargs gaol destroy
}
