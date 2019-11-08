#!/bin/bash
function shutdown()
{
  redis-cli -h maeve.lan PUBLISH sophie "status:online"
  exit 0
}

function startup()
{
  redis-cli -h maeve.lan PUBLISH sophie "status:offline"
  tail -f /dev/null &
  wait $!
}

trap shutdown SIGTERM
trap shutdown SIGKILL

startup;
