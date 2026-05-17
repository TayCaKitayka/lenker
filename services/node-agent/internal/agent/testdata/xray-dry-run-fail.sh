#!/bin/sh
if [ "$1" != "run" ] || [ "$2" != "-test" ] || [ "$3" != "-config" ] || [ -z "$4" ]; then
  echo "fixture invalid invocation" >&2
  exit 64
fi

if [ ! -s "$4" ]; then
  echo "fixture missing candidate config" >&2
  exit 65
fi

echo "Xray dry-run failed: invalid inbound for smoke fixture" >&2
exit 23
