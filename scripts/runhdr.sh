#!/bin/csh

# takes one argument, the directory name holding the bracketed image set
set viewpoint = $1

#�Canon use file suffix JPG not jpg - this fixes it and copies files to a temp folder.

echo $viewpoint
# mkdir -p /tmp/hdrgen/$viewpoint
# mkdir -p /tmp/hdrgen/$viewpoint/exif 
# mkdir -p /tmp/hdrgen/$viewpoint/pic 
# mkdir -p /tmp/hdrgen/$viewpoint/tif
# mkdir -p /tmp/hdrgen/$viewpoint/tmp

set i = 1

echo -n Getting EXIF data $viewpoint ...
foreach image (`ls /tmp/hdrgen/$viewpoint | egrep 'JPG|jpg'`)
	cp /tmp/hdrgen/$viewpoint/$image /tmp/hdrgen/$viewpoint/$i.jpg
	set pth = (/tmp/hdrgen/$viewpoint/$image)
	echo -n .
	exiftool -FileName $pth | awk '{print $4" "}' | tr -d  '"' | tr -d '\n' >> /tmp/hdrgen/$viewpoint/exif/$viewpoint.dat
	exiftool -FocalLength $pth | awk '{print $4" "}' | tr -d '\n' >> /tmp/hdrgen/$viewpoint/exif/$viewpoint.dat
	exiftool -CreateDate $pth | awk '{print $4" "$5" "}' | tr -d '\n' >> /tmp/hdrgen/$viewpoint/exif/$viewpoint.dat
	exiftool -ImageWidth $pth | awk '{print $4" "}' | tr -d '\n' >> /tmp/hdrgen/$viewpoint/exif/$viewpoint.dat
	exiftool -ImageHeight $pth | awk '{print $4" "}' | tr -d '\n' >> /tmp/hdrgen/$viewpoint/exif/$viewpoint.dat
	exiftool -ExposureTime $pth | awk '{print $4" "}' >> /tmp/hdrgen/$viewpoint/exif/$viewpoint.dat
	# ExposureTime doesn't port well into Excel when in format 1/x seconds
	@ i++
end
# DELIVERS exif data to exif/*.csv
# echo $viewpoint > exif/$viewpoint.csv
# echo Filename, Focal Length, CreationDate, CreationTime, PixelWidth,PixelHeight,ExposureTimeSeconds >> exif/$viewpoint.csv
cat /tmp/hdrgen/$viewpoint/exif/$viewpoint.dat | tr " " , >>  /tmp/hdrgen/$viewpoint/exif/$viewpoint.csv
rm /tmp/hdrgen/$viewpoint/exif/*.dat
echo DONE getting exif data.

#this combines the pictures in one single HDR image
#note also the yourcamera.cam is the profile of your camera, either you already have this or it was created by the previous step

echo Starting hdrgen for $viewpoint ...
hdrgen /tmp/hdrgen/$viewpoint/*.jpg -r /tmp/hdrgen/$viewpoint/responseCurve.cam -o /tmp/hdrgen/$viewpoint/pic/$viewpoint.pic
rm /tmp/hdrgen/$viewpoint/*.jpg
echo DONE generating hdr.
	
# this displays a smaller version for quick inspection
# DELIVERS HDR combined images as tif's in the tif folder

echo Resizing image...
# pfilt -1 -x 4000 -y 2800 pic/$viewpoint.comb.pic > pic/$viewpoint.comb.hdr &
pfilt -1 -x 4000 -y 2800 /tmp/hdrgen/$viewpoint/pic/$viewpoint.pic > /tmp/hdrgen/$viewpoint/pic/$viewpoint.hdr 
# pfilt -x 1000 -p 1 pic/$viewpoint.comb.pic > tmp/$viewpoint.comb.hdr &
# pfilt -x 1000 -p 1 /tmp/hdrgen/$viewpoint/pic/$viewpoint.comb.pic > /tmp/hdrgen/$viewpoint/tmp/$viewpoint.comb.hdr 
wait
echo Done resizing image.

echo Creating Tiff...
# rm /tmp/$viewpoint/pic/*.pic
ra_tiff /tmp/hdrgen/$viewpoint/pic/$viewpoint.hdr /tmp/hdrgen/$viewpoint/tif/$viewpoint.tif
echo Done creating Tiff...

convert /tmp/hdrgen/$viewpoint/tif/$viewpoint.tif /tmp/hdrgen/$viewpoint/tif/$viewpoint-base.jpg

# the finished UNSCALED image is in the pic folder as *.comb.hdr
# the finished UNSCALED tif is in the tif folder as *.comb.tif
