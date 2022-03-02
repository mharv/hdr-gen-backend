#!/bin/csh
# Use this script to determine your camera response curve
# Requires a single bracketed set of images in the 'originals' folder
# Delivers camera response curve as 'yourcamera.cam'
# and test image as tif/camera.comb.tif

mkdir -p exif ; mkdir -p pic
rm tmp/*.hdr
rm tmp/*.txt
rm tmp/*.pic

# Canon use file suffix JPG not jpg - this fixes it and copies files to a temp folder.
# Assumes one set of images, one viewpoint
# Single set of images in 'originals' folder
rm tmp/*.jpg
set i = 1
foreach image (`ls originals | egrep 'JPG|jpg'`)
	cp originals/$image tmp/$i.jpg
	@ i++
end
	
# this combines the pictures in one single HDR image
# DELIVERS yourcamera.cam as the profile of your camera

rm pic/camera.comb.hdr
echo hdrgen camera
hdrgen tmp/*.jpg -r yourcamera.cam -o pic/camera.comb.pic
rm tmp/*.jpg
echo unexpose combined image
set expos=`sed -n -e 's/^EXPOSURE=//p' -e '/^$/q' pic/camera.comb.pic | total -p` 
set expos2=`ev 1/$expos`
pfilt -1 -e $expos2 pic/camera.comb.pic > tmp.pic 
mv tmp.pic pic/camera.comb.pic 

# this displays a smaller version for quick inspection
# DELIVERS combined image as tif/camera.comb.tif
	
pfilt -1 -x 4000 -y 2800 pic/camera.comb.pic > pic/camera.comb.hdr &
pfilt -x 1000 -p 1 pic/camera.comb.pic > tmp/camera.comb.hdr &
wait
rm pic/*.pic
ra_tiff tmp/camera.comb.hdr tif/camera.comb.tif

# the finished UNSCALED image is in the pic folder
FINISH:
