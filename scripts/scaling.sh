#!/bin/csh
# Use this script to scale the HDR images to match measured luminance values
# Input is unscaled combined HDR images pic/*.comb.hdr from previous script
# and scale factors input by user for each viewpoint
# For each viewpoint, delivers map of overlaid luminance values as tif/*.vischeck.tif
# and scaled HDR image as pic/*.scaled.hdr
# Once all images are considered to be suitably scaled, continue to next step

# rm tmp/*.*
# how do we read scale factors for each viewpoint and apply them individually by viewpoint?
# input the collection of scale factors here


# takes two arguments, the directory name holding the bracketed image set and new scale factor calculated by backend
set image = $1
set factor = $2
set fullpath = /tmp/hdrgen/$1
# set fullpath = /dev/shm/hdrgen/$1

echo Scaling image $image @ Scale factor = $factor ...
# make a small image to view and test on screen - 2000px x 1400px
# the image will be labelled with 20 x 14 labels = 280 labels

pfilt -1 -x 2000 -y 1400 $fullpath/pic/$image.hdr > $fullpath/tmp/$image.2000.pic
	
# scale this image by scale factor - default is  1  but user may enter alternative
#�test this with the 2000px image
pcomb -h -s $factor $fullpath/tmp/$image.2000.pic > $fullpath/pic/$image.x2000px.pic
wait
echo finished scaling...

		
echo getting luminance values from image $image ...		
# run pvalue and search for any pixel with y,x coordinates = n00,n00
pvalue -o -h -H $fullpath/pic/$image.x2000px.pic | rcalc -e '$1=$1;$2=$2;$3=($3*.265+$4*.67+$5*.065)*179' | awk '{printf "%.0f %.0f %.2f\n", $1, $2, $3}' | tr " " _ | egrep '.00_.00_|..00_..00_|..00_.00_|.00_..00_' | tr _ " " > $fullpath/tmp/shortlist.txt &
wait

echo marking pixels ...		
# this next line makes a little star to mark the point of the sample in the image
psign -cb 0 0 0 -cf 1 0 0 -h 21 "*" > $fullpath/tmp/o.lab
# get the luminance value for each sample point and make a label
echo making labels
foreach y (01 02 03 04 05 06 07 08 09 10 11 12 13)
	foreach x (01 02 03 04 05 06 07 08 09 10 11 12 13 14 15 16 17 18 19)
		set XX = `echo $x | rcalc -e '$1=$1*100'`
		set YY = `echo $y | rcalc -e '$1=$1*100'`
		set lbl = `cat $fullpath/tmp/shortlist.txt | grep ^$XX" "$YY" " | awk '{print $3}'`
		psign -cb 0 0 0 -cf 0.99 0.99 0.99 -h 20 $lbl > $fullpath/tmp/$y$x.lab
	end
end
wait

echo done marking pixels.

echo tiling image ...
# now make a tile measuring 100 x 100 pixels with the sample point and value at top right corner
foreach y (01 02 03 04 05 06 07 08 09 10 11 12 13)
	foreach x (01 02 03 04 05 06 07 08 09 10 11 12 13 14 15 16 17 18 19)
		pcompos -h -b 0 0 0 -x 100 -y 100 =+- $fullpath/tmp/$y$x.lab 89 77 $fullpath/tmp/o.lab 91 83 > $fullpath/tmp/$y$x.hdr &
	end
end
 		
wait
# join all the tiles together to form an array picture
pcompos -h -a 19 -b 1 1 1 $fullpath/tmp/*.hdr > $fullpath/tmp/0000.pic
# and make the array picture the same size as the test picture	
pcompos -h -x 2000 -y 1400 -b 0 0 0 $fullpath/tmp/0000.pic 5 5  > $fullpath/tmp/0000.lab.pic
# superimpose the array onto the test picture
pcomb -h -f addpics.cal $fullpath/tmp/0000.lab.pic $fullpath/pic/$image.x2000px.pic > $fullpath/tmp/check.pic
# make a smaller version to view on screen
pfilt -x /1.7 -y /1.7 $fullpath/tmp/check.pic > $fullpath/tmp/vischeck.pic
# make a smaller falsecolour version to view on screen�and display the image
ra_tiff $fullpath/tmp/vischeck.pic $fullpath/tif/$image.vischeck.tif
# DELIVERS images with overlaid luminance level grid as tif/*.vischeck.tif

convert $fullpath/tif/$image.vischeck.tif $fullpath/tif/$image-scaled.jpg


echo done tiling image.
	
# apply the scale factor to the full size original image - assuming we are happy with result we keep this image.
pcomb -h -s $factor $fullpath/pic/$image.hdr	> $fullpath/tmp/$image.hdr & 
wait
mv $fullpath/tmp/$image.hdr $fullpath/pic/ 

