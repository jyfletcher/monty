#!/bin/bash

export DATE=`date +%Y%m%d`

docker build --squash -t "snmpmon:${DATE}" .
