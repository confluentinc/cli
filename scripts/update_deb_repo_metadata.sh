#!/bin/bash

################
# Prerequisites:
# - Download the current CLI deb repo (exclude .deb files)
# - Use aptly to generate an unsigned deb repo containing the two new .deb packages
# Usage: scripts/update_deb_repo_metadata.sh /path/to/current/deb /path/to/new/deb
################

trap "exit 1" ERR

# The Packages files contain a list of all packages for that architecture.
# Aptly will generate Packages files containing only the new deb packages since we don't download the old ones (to save time and space).
# To update them to include the old information:
# 1. Take the old Packages file and append it to the ends of the new files; this is what Aptly would generate if it had all the packages.
# 2. Recreate the .gz and .bz2 versions of these files.
# 3. For each of MD5, SHA1, SHA256, and SHA512, generate the checksums of Packages, Packages.gz, and Packages.bz2.
# 4. Generate their filesizes (in bytes).
# 5. For each checksum algorithm and file, use sed to replace the checksum:
#    - "^\s\+[a-z0-9]\{${#checksum}\}\s\+[0-9]\+\s\+main/binary-$arch/$file$" this regex matches the line containing a checksum of the same length, a filesize, and the current filename,
#    - "$(printf ' %s %8s %s' $checksum $filesize main/binary-$arch/$file)" replaces the checksum and the filesize, formatted to have 8 spaces in the filesize column (this doesn't practically matter, but it matches the Aptly default style).
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
