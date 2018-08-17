#!/bin/bash

set -eu

src_dir=$(realpath $(dirname $0)/../..)
cd ${src_dir}

tmp_dir=$(mktemp -d)
trap 'cp -f ${tmp_dir}/* ${src_dir}/dist/; rm -rf ${tmp_dir}' EXIT INT QUIT TERM

version=$(awk '$1 == "VERSION" { print $3 }' Makefile)

# Binary buils
while read env; do
    docker run --rm -v ${src_dir}:/root/go/src/facette.io/facette facette/buildenv:${env} make dist-bin
    mv dist/*.tar.gz ${tmp_dir}/
done <<EOF
stretch-amd64
stretch-arm64
stretch-armel
EOF

# Debian packages builds
while read env; do
    [[ -e ${tmp_dir}/facette_${version}.tar.gz ]] && cp ${tmp_dir}/facette_${version}.tar.gz dist/
    docker run --rm -v ${src_dir}:/root/go/src/facette.io/facette facette/buildenv:${env} make dist-deb
    mv dist/*.deb ${tmp_dir}/
done <<EOF
bionic-amd64
stretch-amd64
stretch-arm64
stretch-armel
EOF

# Docker image build
make dist-docker

# vim: ts=4 sw=4 et
