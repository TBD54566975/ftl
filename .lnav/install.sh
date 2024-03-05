#!/usr/bin/env bash

if ! command -v lnav &> /dev/null; then
  echo "error: lnav is not installed. Please install it first."
  exit 1
fi

version="$(lnav --version | cut -d'.' -f2)"
if [[ -z "$version" || "$version" -lt 12 ]]; then
  echo "error: lnav version 0.12 or higher is required"
  exit 1
fi

lnav_format="$(lnav -m format ftl_json source 2> /dev/null)"
if [ -n "$lnav_format" ]; then
  confirm="n"
  # prevent loss of customizations
  #  - https://github.com/tstack/lnav/issues/1240
  echo "warning: ftl_json format already exists '$lnav_format'"
  read -r -p "overwrite? [y/n] " confirm
  if [[ "$confirm" =~ [NnQq] ]]; then
    exit 0
  fi
fi

# always resolve to script directory to install the format
script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" &> /dev/null && pwd)"
find "$script_dir" -name '*.json' -print0 | xargs lnav -i
