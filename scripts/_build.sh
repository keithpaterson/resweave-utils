using colors

run_build_usage() {
  echo "  where $(color -bold op) is:"
  echo "    $(color -lt_green \<empty\>) : build the library"
  echo "    $(color -lt_green test)    : build the library for test"
}

run_build() {
  local _op=$1
  shift
  case ${_op} in
    -h|--help)
      run_handler_usage build
      exit 1
      ;;
    "")
      echo "Build ..."
      go build ./... $*
      ;;
    generate)
      _run test generate $*
      ;;
    test)
      echo "Build for test ..."
      _run test generate
      go build -tags testutils ./... $*
      ;;
    *)
      error "Unrecognized $(color -bold op): '${_op}'"
      exit 1
      ;;
  esac
}
