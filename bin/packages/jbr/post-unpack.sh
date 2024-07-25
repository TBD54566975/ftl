#!/bin/bash
set -xeuo pipefail

mkdir -p "$1/lib/hotswap"

curl -fsSL -o "$1/lib/hotswap/hotswap-agent.jar" \
  https://github.com/HotswapProjects/HotswapAgent/releases/download/RELEASE-1.4.1/hotswap-agent-1.4.1.jar
