#!/bin/bash
export PATH=$JAVA_HOME/bin:$PATH
[ "$MEM" == "" ] && MEM="2G"
[ "$HOST" == "" ] && HOST="localhost"
[ "$PORT" == "" ] && PORT="9669"

OPTS="-server -XX:+UseConcMarkSweepGC -XX:+CMSIncrementalMode -XX:+CMSIncrementalPacing -XX:CMSIncrementalDutyCycleMin=0 -XX:CMSIncrementalDutyCycle=15 -XX:+HeapDumpOnOutOfMemoryError -Xmx$MEM -Duser.timezone=UTC -Dfile.encoding=UTF-8"
$JAVA_HOME/bin/java $OPTS \
  -D-Djetty.host=$HOST -Djetty.port=$PORT \
  -Dsd.webapp.auth=none -Dsd.rest.auth=none \
  -jar streamdrill-launcher-1.1.jar
