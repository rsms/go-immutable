#!/bin/bash
set -e
cd "$(dirname "$0")"

OPT_HELP=false
GO_TEST_ARGS=( -test.v )

# parse args
while [[ $# -gt 0 ]]; do
  case "$1" in
  -h|-help|--help)
    OPT_HELP=true
    shift
    ;;
  -test)
    shift
    if [[ "$1" == "" ]] || [[ "$1" == "-"* ]]; then
      echo "$0: Missing value for option -test" >&2
      OPT_HELP=true
    else
      GO_TEST_ARGS+=( "-run=$1" )
      shift
    fi
    ;;
  -bench)
    shift
    # echo "Benchmarking enabled. Writing profile to bench.pprof."
    # echo "To inspect profile: go tool pprof '$PWD/bench.pprof'"
    # echo "Or start a web server: 'go tool pprof -http 127.0.0.1:8091 bench.pprof'"
    if [[ "$1" != "" ]] && [[ "$1" != "-"* ]]; then
      # use pattern for filtering what benchmarks to run.
      # e.g. -bench Foo
      # GO_TEST_ARGS+=( "-bench=$1" -cpuprofile bench.pprof )
      GO_TEST_ARGS+=( "-bench=$1" )
      shift
    else
      # -benchmem
      # GO_TEST_ARGS+=( -bench=. -cpuprofile bench.pprof )
      GO_TEST_ARGS+=( -bench=. )
    fi
    echo "GO_TEST_ARGS ${GO_TEST_ARGS[@]}"
    ;;
  *)
    echo "$0: Unknown option $1" >&2
    OPT_HELP=true
    shift
    ;;
  esac
done
if $OPT_HELP; then
  echo "usage: $0 [options]" >&2
  echo "options:" >&2
  echo "  -h, -help       Show help" >&2
  echo "  -bench [prefix] Run benchmarks" >&2
  echo "  -test prefix    Run only specific tests (when absent run all tests)" >&2
  exit 1
fi

progpid=
function cleanup {
  kill $progpid
  set +e
  wait $progpid
  set -e
}
trap cleanup EXIT

# make sure we can ctrl-c in the while loop
trap exit SIGINT

if ! (which fswatch >/dev/null); then
  echo "Missing fswatch. See http://emcrisostomo.github.io/fswatch/" >&2
  exit 1
fi

while true; do
  go test "${GO_TEST_ARGS[@]}" &
  progpid=$!
  fswatch -1 -l 0.2 -E --exclude='.*' --include='\.go$' .
  echo "———————————————————— restarting ————————————————————"
  if [[ "$progpid" != "" ]]; then
    set +e
    kill $progpid
    wait $progpid
    set -e
  fi
done
