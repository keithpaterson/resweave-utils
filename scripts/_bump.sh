_bump_verbose=yes

run_bump_usage() {
  echo "  where:"
  echo "    $(color -lt_green -r) <project-root> (required): sets the project root directory."
  echo "  OPTIONS is one or more of:"
  echo "    $(color -lt_green -d)                           : Dry run (don't commit tags to git)"
  echo "    $(color -lt_green -p\|-c\|-v)                     : Print the current version and exit"
  echo "    $(color -lt_green -h)                           : Print this message"
  echo "    $(color -lt_green -o)                           : Omit the alpha portion of the version."
  echo "    $(color -lt_green -s)                           : Silent (no output except final output)"
  echo "    $(color -lt_green -b [alpha\|patch\|minor\|major]) : Bump the appropriate tag $(color -italics [defaults to alpha if not supplied])."
}

run_bump() {
  local _project_root=
  local _version="alpha"
  local _include_alpha=yes
  local _dry_run=yes
  local _print_version=
  local _debug=

  while [ $# -gt 0 ]; do
    case $1 in
      -h|--help)
        run_handler_usage bump
        exit 1
        ;;
      -d|--dry-run)
        _dry_run=yes
        ;;
      -o|--omit-alpha)
        _include_alpha=
        ;;
      -s|--silent)
        _bump_verbose=
        ;;
      -p|-c|-v|--print|--version)
        _print_version=yes
        ;;
      -b|--bump)
        _version=$2
        shift
        ;;
      -r|--project-root)
        _project_root=$2
        shift
        ;;
      --debug)
        # HIDDEN option but useful
        _debug=yes
        ;;
      *)
        error "Unexpected option: $(color -bold -lt_red $1)'"
        echo
        exit 7
        ;;
    esac
    shift
  done

  local _current_version=$(git describe --tags --abbrev=0 2> /dev/null)
  [ -z "${_current_version}" ] && _current_version="v0.0.0"

  if [ -n "${_print_version}" ]; then
    echo ${_current_version}
    exit 0
  fi

  if [ -z "${_project_root}" ]; then
    error "$(color -bold -lt_red \<PROJECT ROOT\>) is required"
    echo
    run_bump_usage
    exit 5
  fi

  _bump_verbose_message Bumping from: ${_current_version}
  local _major=$(echo "${_current_version}" | sed -E "s/^v?([[:digit:]]+)\..*$/\1/g")
  local _minor=$(echo "${_current_version}" | sed -E "s/^v?[[:digit:]]+\.([[:digit:]]+)\..*$/\1/g")
  local _patch=$(echo "${_current_version}" | sed -E "s/^v?[[:digit:]]+\.[[:digit:]]+\.([[:digit:]]+)-?.*$/\1/g")
  local _alpha_date=$(echo "${_current_version}" | sed -E "s/^v?[[:digit:]]+\.[[:digit:]]+\.[[:digit:]]+-alpha-([[:digit:]]+)\..*$/\1/g")
  local _alpha=$(echo "${_current_version}" | sed -E "s/^v?[[:digit:]]+\.[[:digit:]]+\.[[:digit:]]+-alpha-[[:digit:]]+\.([[:digit:]]+).*$/\1/g")

  if [ -n "${_debug}" ]; then
    echo -e "Major Version: ${_major}\nMinor Version: ${_minor}\nPatch Version: ${_patch}\nAlpha Version: ${_alpha}\nAlpha Date: ${_alpha_date}\n"
    echo -e "Assembled    : ${_major}.${_minor}.${_patch}-alpha-${_alpha_date}.${_alpha}"
  fi

  local _new_date=$(date '+%m%d%y')
  case ${_version} in
    major)
      _major=$((_major+1))
      _minor=0
      _patch=0
      _alpha=1
      _alpha_date=${_new_date}
      ;;
    minor)
      _minor=$((_minor+1))
      _patch=0
      _alpha=1
      _alpha_date=${_new_date}
      ;;
    patch)
      _patch=$((_patch+1))
      _alpha=1
      _alpha_date=${_new_date}
      ;;
    alpha)
      if [[ ${_new_date} != ${_alpha_date} ]]; then
        _alpha=1
      else
        _alpha=$((_alpha+1))
      fi
      _alpha_date=${_new_date}
      ;;
    *)
      error "Unexpected bump value $(color -bold -lt_red ${_version})"
      echo
      run_bump_usage
      exit 6
      ;;
  esac

  _full_version="v${_major}.${_minor}.${_patch}"
  [ -n "${_include_alpha}" ] && _full_version+="-alpha-${_alpha_date}.${_alpha}"

  _bump_verbose_message "Bumped      : ${_full_version}"

  if [ -z "${_dry_run}" ]; then
    git tag -a ${_full_version}
    git push --tags
  else
    echo "${_full_version}"
    _bump_verbose_message "git tag -a ${_full_version}"
    _bump_verbose_message "git push --tags"
  fi
}

#
# PRIVATE FUNCTIONS
#

_bump_verbose_message() {
  [ -z "${_bump_verbose}" ] && return

  echo -e $*
}
