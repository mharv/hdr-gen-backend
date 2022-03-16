#!/bin/csh
# Increases exposure of selected image
# Input image input.pic
# DELIVERS new input.pic, over-writing original input.pic
# DELIVERS tif image tif/input.tif

# NEW SCRIPT BELOW
# should only need 2 arguments
echo "UP exposing image " $1 @ factor $2

set fullpath = /tmp/hdrgen/$1
# set fullpath = /dev/shm/hdrgen/$1

# $1 = image name
# $2 = exposure factor

# mkdir -p $fullpath/tmp
# mkdir -p $fullpath/tif

# apply exposure, move to tmp
pfilt -e +$2 $fullpath/pic/$1.hdr > $fullpath/tmp/$1.hdr 
# pfilt -x /5 -y /5 -e +$2 $fullpath/pic/$1.hdr > $fullpath/tmp/$1.resized.hdr 
wait
# create a scaled version for display

# move from tmp back into pic, should overwrite
mv $fullpath/tmp/$1.hdr $fullpath/pic/
# mv tmp/$1.resized.hdr pic/
wait

# remove from tmp if exists

ra_tiff $fullpath/pic/$1.hdr $fullpath/tif/$1.tif
# ra_tiff pic/$1.resized.hdr tif/$1.resized.tif

convert $fullpath/tif/$1.tif $fullpath/tif/$1-exposed.jpg

# example usage
# ./upexpose VP_Euston.comb