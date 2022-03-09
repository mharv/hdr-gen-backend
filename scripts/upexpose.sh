#!/bin/csh
# Increases exposure of selected image
# Input image input.pic
# DELIVERS new input.pic, over-writing original input.pic
# DELIVERS tif image tif/input.tif

# NEW SCRIPT BELOW
# should only need 2 arguments
echo "UP exposing image " $1 @ factor $2

# $1 = image name
# $2 = exposure factor

# mkdir -p /tmp/hdrgen/$1/tmp
# mkdir -p /tmp/hdrgen/$1/tif

# apply exposure, move to tmp
pfilt -e +$2 /tmp/hdrgen/$1/pic/$1.hdr > /tmp/hdrgen/$1/tmp/$1.hdr 
# pfilt -x /5 -y /5 -e +$2 /tmp/hdrgen/$1/pic/$1.hdr > /tmp/hdrgen/$1/tmp/$1.resized.hdr 
wait
# create a scaled version for display

# move from tmp back into pic, should overwrite
mv /tmp/hdrgen/$1/tmp/$1.hdr /tmp/hdrgen/$1/pic/
# mv tmp/$1.resized.hdr pic/
wait

# remove from tmp if exists

ra_tiff /tmp/hdrgen/$1/pic/$1.hdr /tmp/hdrgen/$1/tif/$1.tif
# ra_tiff pic/$1.resized.hdr tif/$1.resized.tif

convert /tmp/hdrgen/$1/tif/$1.tif /tmp/hdrgen/$1/tif/$1.jpg

# example usage
# ./upexpose VP_Euston.comb