# common functions used by the scripts; makes it possible to create standalone operations 
# that execute commands using the existing scripts

# use old syntax for root-script-name since negative indexing started in bash 4.2 and sometimes that's not installed.
_root_script_name=$(basename ${BASH_SOURCE[${#BASH_SOURCE[@]}-1]})
_category_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

_script_common=loaded

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
    echo "$(color -lt_green ${_h}) commands: (./${_root_script_name} ${_h} [op] ...)"
    _run ${_h}_usage
    echo
  )
}

_category_file() {
  if [ -z "$1" ]; then
    echo "ERROR: category_file() requires a category name."
    exit 1
  fi
  echo ${_category_dir}/_${1}.sh
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
[ -z "${_menu_engine}" ] && source ${_category_dir}/_menu_engine.sh

