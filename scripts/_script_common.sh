# common functions used by the scripts; makes it possible to create standalone operations 
# that execute commands using the existing scripts

# The intended use for this is to source this file from your script:
#   ```
#   source /where/this/script/lives/_script_common.sh
#   ```
#
# GLOBAL VARIABLES:
#
# The following variables can be set before sourcing this file in order to control its behaviour:
#   ME_ROOT_SCRIPT_NAME: displayed in help text, this defaults to the name of the base script
#   ME_CATEGORY_DIR:     defines where to locate category subscripts.  Defaults to the directory this file is in.
#   ME_ENGINE:           set to a non-empty string if you don't want/need to load the category handlers

_me_root_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# use old syntax for root-script-name since negative indexing started in bash 4.2 and sometimes that's not installed.
ME_ROOT_SCRIPT_NAME=${ME_ROOT_SCRIPT_NAME:-$(basename ${BASH_SOURCE[${#BASH_SOURCE[@]}-1]})}
ME_CATEGORY_DIR=${ME_CATEGORY_DIR:-${_me_root_dir}}

ME_COMMON=loaded

using() {
  if [ -z "$1" ]; then
    echo "ERROR: using() requires a category name."
    exit 1
  fi
  source $(_category_file ${1})
}

run_handler_usage() {
  local _h=$1
  (
    using ${_h}
    echo "$(color -lt_green ${_h}) commands: (./${ME_ROOT_SCRIPT_NAME} ${_h} [op] ...)"
    _run ${_h}_usage
    echo
  )
}

_category_file() {
  if [ -z "$1" ]; then
    echo "ERROR: category_file() requires a category name."
    exit 1
  fi
  echo ${ME_CATEGORY_DIR}/_${1}.sh
}

# $1: the command plus arguments
# $*: parameters to pass to the command 
#
# Examples:
#   _run clean                => source '_general.sh" and calls "run_general clean"
#   _run build ci             => source '_build.sh' and calls "run_build ci"
#   _run test really special  => source '_test.sh' and calls "run_test really special"
_run() {
  local _cmd=$1
  [[ -z "${_cmd}" ]] && return 2

  if [[ $(type -t run_${_cmd}) != function ]]; then
    local _category=$1
    local _file=$(_category_file ${_category})
    if [ ! -f ${_file} ]; then
      _category=general
      _cmd="general ${_cmd}"
    fi
    using ${_category}
  fi
  shift
  _run_command $_cmd $*
}

# search for 'run_$1' and execute it
# e.g. _run_command test => calls 'run_test()'
#      _run_command build_go -x -y -z => calls 'run_build_go()' with parameters -x -y -z
_run_command() {
  local _cmd="run_${1/-/_}"
  shift
  if [[ $(type -t ${_cmd}) == function ]]; then
    ${_cmd} $@
  else
    echo no such command "${_cmd}"
    return 2
  fi
}

# load the menu engine after defining things it will need.  REQUIRES that the engine be in the same folder as this file.
# You can avoid loading the engine scripts by setting ${_menu_engine} prior to sourcing _script_common.sh
# e.g.
# ```
# _menu_engine=excluded
# source /where/this/script/lives/_script_common.sh
# ```
[ -z "${ME_ENGINE}" ] && source ${_me_root_dir}/_menu_engine.sh

