#!/bin/sh

tmpfile=$(mktemp)
installfile=/usr/local/bin/tubectl
os=$(uname)

echo

case $os in
  Darwin)
    echo "Installing for OS X..."
    curl -Lo $tmpfile https://github.com/Freshly/tuber/releases/download/v1.0/tuber_macos &> /dev/null
    mv $tmpfile $installfile
    chmod +x $installfile
    ;;
  Linux)
    echo "Installing for Linux..."
    curl -Lo $tmpfile https://github.com/Freshly/tuber/releases/download/v1.0/tuber_linux &> /dev/null
    mv $tmpfile $installfile
    chmod +x $installfile
    ;;
  *)
    echo "Sorry Jordan, we don't support Windows."
    exit 1
    ;;
esac
