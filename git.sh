#!/bin/bash
if [ ! "$1" ]; then
  git add . && git commit -m "$(uname -s) $(date "+%Y-%m-%d %H:%M:%S")" && git push
else
  git add . && git commit -m "$1" && git push
fi
