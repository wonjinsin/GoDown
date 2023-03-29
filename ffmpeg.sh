#!/bin/bash

cd $1; 
find *.$3 | sort -n | sed 's:\\ :\\\\\\ :g'| sed 's/^/file /' > fl.txt; 
err=`ffmpeg -v error -f concat -i fl.txt -c copy -y $2.mp4 2>&1 | grep -E "Impossible|Invalid"`;
if [ "$err" == "" ]; then
	rm fl.txt *.$3;
fi