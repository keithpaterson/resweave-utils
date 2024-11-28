using colors

_test_report_dir=${_root_dir}/.reports

## original
run_test_usage() {
  echo "  where $(color -bold op) is:"
  echo "    $(color -lt_green \<empty\>)  :  run unit tests"
  echo "    $(color -lt_green coverage) : run unit tests and generate coverage"
  echo "    $(color -lt_green generate) : generate mocks"
  echo "      coverage reports will be found in '${_test_report_dir}'"
}

run_test() {
  local _op=$1
  shift

  case ${_op} in
    "")
      _unit_tests $*
      ;;
    coverage)
      _unit_coverage $*
      ;;
    generate|mocks)
      _generate_mocks $*
      ;;
    -h|--help)
      run_handler_usage test
      exit 1
      ;;
    *)
      error "unrecognized op: $(color -bold -lt_red $_op)"
      ;;
  esac
}

run_linter() {
  staticcheck -tags testutils -f stylish ./...
}

_unit_tests() {
  _run build test
  echo "Running unit tests..."
  mkdir -p ${_test_report_dir}
  _test_check_and_install_ginkgo
  ginkgo --tags testutils --repeat 1 -r --output-dir ${_test_report_dir} --json-report unit_tests.json ./... > ${_test_report_dir}/unit_tests.log 2>&1
  local _result=$?
  cat ${_test_report_dir}/unit_tests.log
  if [ ${_result} -ne 0 ]; then
      exit ${_result}
  fi
}

_unit_coverage() {
  _run build test
  echo "generating test coverage..."
  mkdir -p ${_test_report_dir}
  rm -f ${_test_report_dir}/coverage.raw.out ${_test_report_dir}/coverage.out
  go test --tags testutils --test.coverprofile ${_test_report_dir}/coverage.raw.out ./... | grep -v mocks 
  local _result=$?

  # filter out mocks directories from coverage
  grep -vE 'mocks/|utility/test/' ${_test_report_dir}/coverage.raw.out > ${_test_report_dir}/coverage.out
  go tool cover -html=${_test_report_dir}/coverage.out -o ${_test_report_dir}/coverage.html
  if [ ${_result} -ne 0 ]; then
      exit ${_result}
  fi
}

_generate_mocks() {
  go generate $* ./...
}

_test_check_and_install_ginkgo() {
  if ! command -v ginkgo &> /dev/null; then
    echo install ginkgo
    go install github.com/onsi/ginkgo/v2/ginkgo
  fi
}
