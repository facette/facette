#!/bin/bash

set -eu

cd $(dirname $0)/../..

version=$(awk '$1 == "VERSION" { print $3 }' Makefile)
if [[ "${version}" =~ (.+)dev$ ]]; then
    debian_version="${BASH_REMATCH[1]}-0~git$(date +%Y%m%d).$(git rev-parse --short HEAD)"
elif [[ "${version}" =~ (.+)(alpha|beta|dev|rc)([0-9]+)$ ]]; then
    debian_version="${BASH_REMATCH[1]}-1~${BASH_REMATCH[2]}${BASH_REMATCH[3]}"
else
    debian_version="${version}-1"
fi

target_arch=$(dpkg-architecture -q DEB_TARGET_ARCH)
target_suite=$(lsb_release -sc)

# Generate source archive
[[ ! -e dist/facette_${version}.tar.gz ]] && make dist-source
cp dist/facette_${version}.tar.gz ../facette_${debian_version%-*}.orig.tar.gz

# Generate Debian changelog file from template
sed \
    -e "s/%%VERSION%%/${version}/g" \
    -e "s/%%DEBIAN_VERSION%%/${debian_version}/g" \
    -e "s/%%DATE%%/$(date -R)/g" \
    debian/changelog.tmpl >debian/changelog

# Build package
DEB_BUILD_OPTIONS=nocheck debuild \
    --preserve-envvar CC \
    --preserve-envvar CGO_ENABLED \
    --preserve-envvar GOARCH \
    --preserve-envvar GOARM \
    --preserve-envvar GOOS \
    --preserve-envvar PATH \
    -a${target_arch} -d -uc -us

cp ../facette_${debian_version}_${target_arch}.deb dist/facette_${version}_${target_suite}-${target_arch}.deb

# vim: ts=4 sw=4 et
