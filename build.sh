#!/usr/bin/env bash

_root_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
_gitroot_dir=$(cd ${_root_dir} && git rev-parse --show-toplevel)

# move into the root folder to run everything else
cd $_root_dir
source scripts/_menu_engine.sh
