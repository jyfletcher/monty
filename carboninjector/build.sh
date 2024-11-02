#!/bin/bash

DATE=`date +%Y%m%d`

docker build --squash -t "carboninjector:${DATE}" .


