using colors

_lint_report_dir=${_root_dir}/.reports
_lint_report_file=${_lint_report_dir}/lint-report.json

run_lint_usage() {
  echo "  where $(color -bold op) is:"
  echo "    $(color -lt_green \<empty\>) : run the linter"
  echo "    $(color -lt_green report)  : run the linter and generate a json report"
  echo "      lint reports will be found in '${_lint_report_dir}'"
}

run_lint() {
  local _op=$1
  shift

  case ${_op} in
    "")
      _lint
      ;;
    report|-r|--report)
      _lint json
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

_lint() {
  local _tags="-tags testutils"

  if [ -z "${1}" ]; then
    staticcheck ${_tags} -f stylish ./...
  else
    _ensure_reports_dir
    staticcheck ${_tags} -f json ./... > ${_lint_report_file}
  fi
}

_ensure_reports_dir() {
  mkdir -p ${_lint_report_dir}
}