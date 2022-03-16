#!/bin/csh

# takes one argument, the directory name holding the bracketed image set
set viewpoint = $1

set fullpath = /tmp/hdrgen/$viewpoint
# set fullpath = /dev/shm/hdrgen/$1

#�Canon use file suffix JPG not jpg - this fixes it and copies files to a temp folder.

echo $viewpoint
# mkdir -p $fullpath
# mkdir -p $fullpath/exif 
# mkdir -p $fullpath/pic 
# mkdir -p $fullpath/tif
# mkdir -p $fullpath/tmp

set i = 1

echo -n Getting EXIF data $viewpoint ...
foreach image (`ls $fullpath | egrep 'JPG|jpg'`)
	cp $fullpath/$image $fullpath/$i.jpg
	set pth = ($fullpath/$image)
	echo -n .
	exiftool -FileName $pth | awk '{print $4" "}' | tr -d  '"' | tr -d '\n' >> $fullpath/exif/$viewpoint.dat
	exiftool -FocalLength $pth | awk '{print $4" "}' | tr -d '\n' >> $fullpath/exif/$viewpoint.dat
	exiftool -CreateDate $pth | awk '{print $4" "$5" "}' | tr -d '\n' >> $fullpath/exif/$viewpoint.dat
	exiftool -ImageWidth $pth | awk '{print $4" "}' | tr -d '\n' >> $fullpath/exif/$viewpoint.dat
	exiftool -ImageHeight $pth | awk '{print $4" "}' | tr -d '\n' >> $fullpath/exif/$viewpoint.dat
	exiftool -ExposureTime $pth | awk '{print $4" "}' >> $fullpath/exif/$viewpoint.dat
	# ExposureTime doesn't port well into Excel when in format 1/x seconds
	@ i++
end
# DELIVERS exif data to exif/*.csv
# echo $viewpoint > exif/$viewpoint.csv
# echo Filename, Focal Length, CreationDate, CreationTime, PixelWidth,PixelHeight,ExposureTimeSeconds >> exif/$viewpoint.csv
cat $fullpath/exif/$viewpoint.dat | tr " " , >>  $fullpath/exif/$viewpoint.csv
rm $fullpath/exif/*.dat
echo DONE getting exif data.

#this combines the pictures in one single HDR image
#note also the yourcamera.cam is the profile of your camera, either you already have this or it was created by the previous step

echo Starting hdrgen for $viewpoint ...
hdrgen $fullpath/*.jpg -r $fullpath/responseCurve.cam -o $fullpath/pic/$viewpoint.pic
rm $fullpath/*.jpg
echo DONE generating hdr.
	
# this displays a smaller version for quick inspection
# DELIVERS HDR combined images as tif's in the tif folder

echo Resizing image...
# pfilt -1 -x 4000 -y 2800 pic/$viewpoint.comb.pic > pic/$viewpoint.comb.hdr &
pfilt -1 -x 4000 -y 2800 $fullpath/pic/$viewpoint.pic > $fullpath/pic/$viewpoint.hdr 
# pfilt -x 1000 -p 1 pic/$viewpoint.comb.pic > tmp/$viewpoint.comb.hdr &
# pfilt -x 1000 -p 1 $fullpath/pic/$viewpoint.comb.pic > $fullpath/tmp/$viewpoint.comb.hdr 
wait
echo Done resizing image.

echo Creating Tiff...
# rm /tmp/$viewpoint/pic/*.pic
ra_tiff $fullpath/pic/$viewpoint.hdr $fullpath/tif/$viewpoint.tif
echo Done creating Tiff...

convert $fullpath/tif/$viewpoint.tif $fullpath/tif/$viewpoint-base.jpg

# the finished UNSCALED image is in the pic folder as *.comb.hdr
# the finished UNSCALED tif is in the tif folder as *.comb.tif
