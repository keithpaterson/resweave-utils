#!/usr/bin/env bash

_script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
_root_dir=${_script_dir}
_test_report_dir=${_root_dir}/.reports

show_usage() {
    echo "build.sh test [coverage]"
    echo "  runs all the tests"
    echo "  'coverage' generates a code-coverage report"
}

run_tests() {
    local _op=$1
    shift

    case ${_op} in
      coverage)
        _unit_coverage $*
        ;;
      *)
        _unit_tests $*
        
    esac
    go test --tags=testutils ./...
}

_unit_tests() {
  echo "Running unit tests..."
  mkdir -p ${_test_report_dir}
  _test_check_and_install_ginkgo
  ginkgo --tags testutils --repeat 1 -r --output-dir .reports --json-report unit_tests.json ./... > .reports/unit_tests.log 2>&1
  local _result=$?
  cat .reports/unit_tests.log
  if [ ${_result} -ne 0 ]; then
      exit ${_result}
  fi
}

_unit_coverage() {
    echo "generating test coverage..."
    mkdir -p ${_test_report_dir}
    rm -f .reports/coverage.raw.out .reports/coverage.out
    go test --tags testutils --test.coverprofile .reports/coverage.raw.out ./... | grep -v mocks 
    local _result=$?

    # filter out mocks directories from coverage
    grep -vE 'mocks/|utility/test/' .reports/coverage.raw.out > .reports/coverage.out
    go tool cover -html=.reports/coverage.out -o .reports/coverage.html
    if [ ${_result} -ne 0 ]; then
        exit ${_result}
    fi
}

_test_check_and_install_ginkgo() {
  if ! command -v ginkgo &> /dev/null; then
    echo install ginkgo
    go install github.com/onsi/ginkgo/v2/ginkgo
  fi
}

#
# MAIN
#

while [ $# -gt 0 ]; do
  _op=$1
  shift

  case ${_op} in
    -h|--help|help)
      show_usage
      exit 0
      ;;
    test)
      run_tests $*
      exit 0
      ;;
    *)
      echo "ERROR: unexpected: '${_op}'"
      exit 1
      ;;
  esac
done
