#!/bin/bash

if [ $# -lt 1 ]
then
	echo "Usage $0 <version>"
	exit 1
fi

VER=$1
DESC="Remote network packet capturing tool"
FPM_IMAGE="tdimitrov/tranqap-pkg:latest"

for pkg_type in "deb" "rpm"
do
	docker run --rm -v `pwd`:/pkg ${FPM_IMAGE} -s dir -t ${pkg_type} -n tranqap -v "${VER}" --description "${DESC}" --license "GPL-3.0" --url "https://github.com/tdimitrov/tranqap" ./tranqap=/usr/bin/
done
