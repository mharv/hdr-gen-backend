#!/bin/csh
# Use this script to create a map of luminance values overlaid on the HDR image created by the previous script
# Input is unscaled combined HDR images pic/*.comb.hdr from previous script
# For each viewpoint, delivers map of overlaid luminance values as tif/*.vischeck.tif
# and scaled HDR image as pic/*.scaled.hdr


# mkdir -p /tmp/hdrgen/$1/pic
# mkdir -p /tmp/hdrgen/$1/tif
# mkdir -p /tmp/hdrgen/$1/tmp

# takes one argument, the hdr image name without extension 
set image = $1
set factor = 1

echo Scaling image $image @ Scale factor = $factor ...
# make a small image to view and test on screen - 2000px x 1400px
# the image will be labelled with 20 x 14 labels = 280 labels

pfilt -1 -x 2000 -y 1400 /tmp/hdrgen/$image/pic/$image.hdr > /tmp/hdrgen/$image/tmp/$image.2000.pic
	
# scale this image by scale factor - default is  1  but user may enter alternative
# test this with the 2000px image
pcomb -h -s $factor /tmp/hdrgen/$image/tmp/$image.2000.pic > /tmp/hdrgen/$image/pic/$image.x2000px.pic
wait
echo finished scaling...

		
echo getting luminance values from image $image ...		
# run pvalue and search for any pixel with y,x coordinates = n00,n00
pvalue -o -h -H /tmp/hdrgen/$image/pic/$image.x2000px.pic | rcalc -e '$1=$1;$2=$2;$3=($3*.265+$4*.67+$5*.065)*179' | awk '{printf "%.0f %.0f %.2f\n", $1, $2, $3}' | tr " " _ | egrep '.00_.00_|..00_..00_|..00_.00_|.00_..00_' | tr _ " " > /tmp/hdrgen/$image/tmp/shortlist.txt &
wait

echo marking pixels ...		
# this next line makes a little star to mark the point of the sample in the image
psign -cb 0 0 0 -cf 1 0 0 -h 21 "*" > /tmp/hdrgen/$image/tmp/o.lab
# get the luminance value for each sample point and make a label
echo making labels
foreach y (01 02 03 04 05 06 07 08 09 10 11 12 13)
	foreach x (01 02 03 04 05 06 07 08 09 10 11 12 13 14 15 16 17 18 19)
		set XX = `echo $x | rcalc -e '$1=$1*100'`
		set YY = `echo $y | rcalc -e '$1=$1*100'`
		set lbl = `cat /tmp/hdrgen/$image/tmp/shortlist.txt | grep ^$XX" "$YY" " | awk '{print $3}'`
		psign -cb 0 0 0 -cf 0.99 0.99 0.99 -h 20 $lbl > /tmp/hdrgen/$image/tmp/$y$x.lab
	end
end
wait

echo done marking pixels.

echo tiling image ...
# now make a tile measuring 100 x 100 pixels with the sample point and value at top right corner
foreach y (01 02 03 04 05 06 07 08 09 10 11 12 13)
	foreach x (01 02 03 04 05 06 07 08 09 10 11 12 13 14 15 16 17 18 19)
		pcompos -h -b 0 0 0 -x 100 -y 100 =+- /tmp/hdrgen/$image/tmp/$y$x.lab 89 77 /tmp/hdrgen/$image/tmp/o.lab 91 83 > /tmp/hdrgen/$image/tmp/$y$x.hdr &
	end
end
 		
wait
# join all the tiles together to form an array picture
pcompos -h -a 19 -b 1 1 1 /tmp/hdrgen/$image/tmp/*.hdr > /tmp/hdrgen/$image/tmp/0000.pic
# and make the array picture the same size as the test picture	
pcompos -h -x 2000 -y 1400 -b 0 0 0 /tmp/hdrgen/$image/tmp/0000.pic 5 5  > /tmp/hdrgen/$image/tmp/0000.lab.pic
# superimpose the array onto the test picture
pcomb -h -f addpics.cal /tmp/hdrgen/$image/tmp/0000.lab.pic /tmp/hdrgen/$image/pic/$image.x2000px.pic > /tmp/hdrgen/$image/tmp/check.pic
# make a smaller version to view on screen
pfilt -x /1.7 -y /1.7 /tmp/hdrgen/$image/tmp/check.pic > /tmp/hdrgen/$image/tmp/vischeck.pic
# make a smaller falsecolour version to view on screen�and display the image
ra_tiff /tmp/hdrgen/$image/tmp/vischeck.pic /tmp/hdrgen/$image/tif/$image.vischeck.tif
# DELIVERS images with overlaid luminance level grid as tif/*.vischeck.tif

convert /tmp/hdrgen/$image/tif/$image.vischeck.tif /tmp/hdrgen/$image/tif/$image-scaled.jpg

# remove temporary files
# rm tmp/*.hdr
# rm tmp/*.lab
# rm tmp/*.pic

echo done tiling image.
	
# apply the scale factor to the full size original image - assuming we are happy with result we keep this image.
pcomb -h -s $factor /tmp/hdrgen/$image/pic/$image.hdr > /tmp/hdrgen/$image/tmp/$image.hdr &
wait
mv /tmp/hdrgen/$image/tmp/$image.hdr /tmp/hdrgen/$image/pic/ 

# rm /tmp/hdrgen/$image/pic/$image.x2000px.pic
# rm /tmp/hdrgen/$image/tmp/shortlist.txt

