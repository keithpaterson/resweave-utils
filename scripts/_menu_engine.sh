#!/usr/bin/env bash

# The intended use for this is to source this file from your script:
#   ```
#   source /where/this/script/lives/_menu_engine.sh
#   ```
#
# Note that this script relies on _script_common.sh and won't work without it.
# In particular, it uses the ${_category_dir} to locate the subscripts.
#
# This script, when sourced, will automatically try to load _script_common.sh
# from the same folder where this file lives.

# the flow for deciding where to locate a function is:
#   if a function `run_$1()` exists, call it with $*
#   if a file `${_category_dir}/_$1.sh` exists, `source` it
#     then call `run_$1()` with $*
#
# each category-handler subscript is expected to:
#   - be found at `${_category_dir}/_<category>.sh`
#   - implement a help function `run_<category>_usage()`
#   - implement an entry point function `run_<category>()`
#
# there may be categories that don't have run entry points but do have functions
# that make life easier (e.g. for docker calls).  These are expected to:
#   - be found at `${_category_dir}/_<category>.sh`
#   - be included by other scripts via `using <name>`
#     where <name> indicates the category; these are included as described above.
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
