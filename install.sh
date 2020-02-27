#!/bin/sh

tmpfile=$(mktemp)
installfile=/usr/local/bin/tubectl
os=$(uname)

case $os in
  Darwin)
    echo "Installing for OS X..."
    curl -Lo $tmpfile https://github.com/Freshly/tuber/releases/download/v1.0/tuber_macos
    mv $tmpfile $installfile
    chmod +x $installfile
    break
    ;;
  Linux)
    echo "Installing for Linux..."
    curl -Lo $tmpfile https://github.com/Freshly/tuber/releases/download/v1.0/tuber_macos
    mv $tmpfile $installfile
    chmod +x $installfile
    break
    ;;
  *)
    echo "Sorry Jordan, we don't support Windows."
    exit 1
    ;;
esac
