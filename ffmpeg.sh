#!/bin/bash

cd $1; 
find *.ts | sort -n | sed 's:\\ :\\\\\\ :g'| sed 's/^/file /' > fl.txt; 
ffmpeg -err_detect ignore_err -f concat -i fl.txt -c copy $2.mp4; 
rm fl.txt *.ts
