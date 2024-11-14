#!/usr/bin/env bash

# The intended use for this is
#   ./agt command (for commands found in '_general.sh')
#   ./agt category command
# where
#   command is a function to run or a 'general' function
#   category indicates a group of functions
#
# for example:
#   ./agt clean
#     loads the general category and runs 'run_general clean'
#   ./agt test integration
#     loads the test category and runs a function with parameters: 'run_test integration'
#
# the flow for deciding where to locate a function is:
#   if a function `run_$1()` exists, call it with $*
#   if a file `scripts/_$1.sh` exists, `source` it
#     then call `run_$1()` with $*
#
# each category-handler script is expected to:
#   be found at `scripts/_<category>.sh`
#   implement a help function `run_<category>_usage()`
#   implement an entry point function `run_<category>()`
#
# there may be categories that don't have entry points but do have functions
# that make life easier (e.g. for docker calls).  These are expected to:
#   be found at `scripts/_<category>.sh`
#   be included by other scripts via `using <name>`
#     where <name> indicates the category; these are included via `source scripts/_<name>.sh`
#     (the `using()` function is a glorified wrapper around `source`)

_menu_engine=initializing
if [ -z "${_script_common}" ]; then
  # This only works if _script_common.sh is in the same place as this file.
  _local_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
  source ${_local_dir}/_script_common.sh
fi

_menu_engine=loaded
using colors

_handlers=
_all_scripts=

_load_handlers() {
  _all_scripts=$(
    cd ${_category_dir}
    for handler in _*.sh; do
      handler=${handler%.sh}
      echo ${handler/_/}
    done
  )

  for h in ${_all_scripts}; do
    # if a category file has a usage function, store it as a handler
    if grep -E "^run_${h}_usage\(\)" $(_category_file $h) > /dev/null; then
      _handlers+=" ${h}"
    fi
  done
  # convert newline to space and remove leading spaces
  _handlers=${_handlers//$'\n'/ }
  _handlers=${_handlers# }
}

run_usage() {
  echo "$(color -bold ${_root_script_name}) usage:"
  echo "  $(color -bold ./${_root_script_name}) $(color -lt_green \<command\>) [<options>]"
  echo
  for _h in ${_handlers}; do
    run_handler_usage ${_h}
  done
}

_load_handlers

if [ $# -lt 1 ]; then
  run_usage
  exit 1
fi

case $1 in
  -h|--help|help)
  run_usage $*
  exit 0
esac

_run ${CMD} $*
