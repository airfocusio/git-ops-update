#!/bin/bash
set -eo pipefail

if [[ "$OSTYPE" == "linux-gnu"* ]]; then
  VARIANT="linux-amd64"
  mkdir -p "$HOME/bin"
  TARGET="$HOME/bin/git-ops-update"
elif [[ "$OSTYPE" == "darwin"* ]]; then
  VARIANT="darwin-amd64"
  mkdir -p "$HOME/bin"
  TARGET="$HOME/bin/git-ops-update"
else
  echo "Unknown OS type $OSTYPE"
  exit 1
fi
LATEST_VERSION="$(curl -s https://api.github.com/repos/choffmeister/git-ops-update/releases/latest | grep "tag_name" | awk '{print substr($2, 2, length($2)-3)}')"

curl -fsSL -o "$TARGET" "https://github.com/choffmeister/git-ops-update/releases/download/$LATEST_VERSION/git-ops-update-$VARIANT"
chmod +x "$TARGET"
