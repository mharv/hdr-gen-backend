#!/bin/csh
# Use this script to post-process the scaled HDR images
# Input is scaled HDR images from previous script as pic/*.scaled.hdr
# Delivers image applying the human contrast sensitivity function as tif/*/*.s.tif
# and a series of falsecolour maps as tif/*/*.fc.*.tif
# and down-sized scaled HDR images as tmp/*.scaled.hdr
# and re-exposed scaled HDR image as tif/*/*.tif
# the last two can be re-exposed using scripts upexpose or downexpose


# take one argument, image name
set image = $1
set fullpath = /tmp/hdrgen/$1


#### establish a suitable exposure for the images
#�open an empty file to store exposure data

# echo -n tmp/opti.txt
echo getting p values for $1 ...
# shrink each image to 20 x 14 pixels and then use pvalue  - this will give average luminance of each image. The get average of all viewpoint images.
#�use this to set an optimum scale for all images


pfilt -1 -x 20 -y 14 $fullpath/pic/$image.hdr > $fullpath/tmp/$image.20x14.pic
pvalue -o -h -H $fullpath/tmp/$image.20x14.pic | rcalc -e '$1=$1;$2=$2;$3=($3*.265+$4*.67+$5*.065)*179' | awk '{printf "%.3f\n", $3}' | tr "\t" " " | total -m >>  $fullpath/tmp/opti.txt
rm $fullpath/tmp/$image.20x14.pic


#�use this to set an optimum scale - uses 8 x average
set opti = `cat $fullpath/tmp/opti.txt | total -m | rcalc -e '$1=$1*8'`
rm $fullpath/tmp/opti.txt

# now sort out the exposure for each image
# default is +1
set exp = (+1)

echo p values retrieved.

echo creating false colour image for $image ...

#�create image using human contrast sensitivity function
pcond -s $fullpath/pic/$image.hdr | ra_tiff -z -  $fullpath/tif/$image.s.tif &
wait



#�establish a legend size for the falsecolour images, based on total image size
set dims = `getinfo -d $fullpath/pic/$image.hdr`
set lh = `echo $dims[3] | rcalc -e '$1=$1/4'` 
set lw = `echo $dims[3] | rcalc -e '$1=$1/8'` 
set fclw = `echo $dims[3] | rcalc -e '$1=($1/8)+4000'` # legend size for falsecolour - use this later when assembling arrays

# Create falsecolour maps

falsecolor -ip	$fullpath/pic/$image.hdr -s $opti	-lh $lh -lw $lw -n 10 -log 2 | ra_tiff - $fullpath/tif/$image.fc.optm.tif &
falsecolor -ip	$fullpath/pic/$image.hdr -s $opti	-lh $lh -lw $lw -n 10 -log 2 > $fullpath/tmp/$image.fc.optm.pic &
falsecolor -ip	$fullpath/pic/$image.hdr -s 1 		-lh $lh -lw $lw -n 10 -log 4 | ra_tiff - $fullpath/tif/$image.fc.1.tif &
falsecolor -ip	$fullpath/pic/$image.hdr -s 100 		-lh $lh -lw $lw -n 10 -log 4 | ra_tiff - $fullpath/tif/$image.fc.100.tif &
falsecolor -ip	$fullpath/pic/$image.hdr -s 10000 	-lh $lh -lw $lw -n 10 -log 4 | ra_tiff - $fullpath/tif/$image.fc.10000.tif &
# sleep 10 ; echo ; echo this may take a while...; echo
echo falsecolour generation complete.

# set image exposure and display preview
echo setting exposure for $1 ...
pfilt -x /5 -y /5 -e $exp $fullpath/pic/$image.hdr > $fullpath/tmp/$image.hdr
ra_tiff $fullpath/tmp/$image.hdr $fullpath/tif/$image.tif
pfilt -e $exp 	$fullpath/pic/$image.hdr | ra_tiff -z -  $fullpath/tif/$image.tif &
wait


convert $fullpath/tif/$image.fc.100.tif $fullpath/tif/$image-falseColor.jpg
echo processing complete.
