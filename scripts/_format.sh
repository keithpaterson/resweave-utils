using colors

run_format_usage() {
  echo "  where $(color -bold op) is:"
  echo "    $(color -lt_green \<empty\>) :  format source files"
}

run_format() {
  local _op=$1
  shift
  case ${_op} in
    "")
      echo "Format source files ..."
      go fmt $* ./...
      ;;
    *)
      error "Unrecognized $(color -bold op): '${_op}'"
      exit 1
      ;;
  esac
}