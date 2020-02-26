#!/bin/sh

tmpfile=$(mktemp)
installfile=/usr/local/bin/tubectl

curl -Lo $tmpfile https://github.com/Freshly/tuber/releases/download/v1.0/tuber_macos

mv $tmpfile $installfile
chmod +x $installfile
