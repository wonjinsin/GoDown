#!/bin/bash
fl="fl.txt"
msg="No such|Impossible|Invalid"
cd $1; 
find *.$3 | sort -n | sed 's:\\ :\\\\\\ :g'| sed 's/^/file /' > $fl; 
err=`ffmpeg -v error -f concat -i $fl -c copy -y $2.mp4 2>&1 | grep -E "$msg"`;
if [ "$err" == "" ]; then
	rm $fl *.$3;
	exit 0;
fi
exit 1;
