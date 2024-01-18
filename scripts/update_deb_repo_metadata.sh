#!/bin/bash

################
# Prerequisites:
# - Download the current CLI deb repo (exclude .deb files)
# - Use aptly to generate an unsigned deb repo containing the two new .deb packages
# Usage: scripts/update_deb_repo_metadata.sh /path/to/current/deb /path/to/new/deb
################

trap "exit 1" ERR

for arch in amd64 arm64; do
    if [[ -f $1/dists/stable/main/binary-$arch/Packages ]]; then
        cat $1/dists/stable/main/binary-$arch/Packages >> $2/dists/stable/main/binary-$arch/Packages
        rm -f $2/dists/stable/main/binary-$arch/Packages.gz $2/dists/stable/main/binary-$arch/Packages.bz2
        gzip -k -n $2/dists/stable/main/binary-$arch/Packages
        bzip2 -k $2/dists/stable/main/binary-$arch/Packages
        for file in Packages Packages.gz Packages.bz2; do
            for algo in md5sum sha1sum sha256sum sha512sum; do
                checksum=$($algo $2/dists/stable/main/binary-$arch/$file | cut -d ' ' -f 1)
                filesize=$(stat -c %s $2/dists/stable/main/binary-$arch/$file)
                sed -i.bak "s|^\s\+[a-z0-9]\{${#checksum}\}\s\+[0-9]\+\s\+main/binary-$arch/$file$|$(printf ' %s %8s %s' $checksum $filesize main/binary-$arch/$file)|g" $2/dists/stable/Release
                rm -f $2/dists/stable/Release.bak
            done
        done
    fi
done
