#!/bin/bash

read -d '' DEPENDS <<EOF
git  https://github.com/facette/natsort     master
git  https://github.com/fatih/set           master
git  https://github.com/influxdb/influxdb   master  /client /LICENSE
git  https://github.com/nu7hatch/gouuid     master
git  https://github.com/stretchr/powerwalk  master
git  https://github.com/ziutek/rrd          master
EOF

fetch_git() {
	name=$1
	url=$2
	branch=$3

	if [[ ! -d "$SRC_DIR/checkouts/$name" ]]; then
		echo "Fetching $name..."
		mkdir -p $SRC_DIR/checkouts/$name
		git clone --quiet $url $SRC_DIR/checkouts/$name -b $branch
	else
		echo "Updating $name..."
		git -C $SRC_DIR/checkouts/$name pull --quiet
	fi
}

fetch_hg() {
	name=$1
	url=$2
	branch=$3

	if [[ ! -d "$SRC_DIR/checkouts/$name" ]]; then
		echo "Fetching $name..."
		mkdir -p $SRC_DIR/checkouts/$name
		hg clone --quiet $url $SRC_DIR/checkouts/$name -b $branch
	else
		echo "Updating $name..."
		hg -R $SRC_DIR/checkouts/$name pull --quiet
		hg -R $SRC_DIR/checkouts/$name update --quiet
	fi
}

rewrite_imports() {
	name=$1

	echo "Rewriting $name import paths..."
	find $SRC_DIR -type f ! -path "$SRC_DIR/checkouts/*" -name '*.go' \
		-exec sed -e "s@\"$name@\"github.com/facette/facette/thirdparty/$name@" -i {} \;
}

print_usage() {
	echo "Usage: $(basename $0) [-u]" >&2
	exit 1
}

# Parse command-line arguments
[[ $# -ne 0 ]] && print_usage

SRC_DIR=$(dirname $0)

# Fetch dependencies
declare -a PACKAGES

IFS=$'\n'; for entry in $DEPENDS; do
	unset IFS
	read type url branch paths<<<$(echo $entry)

	# Get package name
	name=${url#http*://}

	# Fetch project source and create a local copy
	fetch_$type $name $url $branch

	rm -rf $SRC_DIR/$name/*

	[[ -d "$SRC_DIR/$name" ]] || mkdir -p $SRC_DIR/${name}

	# If specific paths are specified
	if [[ -n "$paths" ]]; then
		for path in $paths; do
			dirname=$(dirname $SRC_DIR/${name}${path})
			[[ -d "$dirname" ]] || mkdir -p $dirname
			cp -a $SRC_DIR/checkouts/${name}${path} $SRC_DIR/${name}${path}
		done
	else
		cp -a $SRC_DIR/checkouts/${name}/* $SRC_DIR/${name}

	fi

	PACKAGES+=(${name})
done

for package in ${PACKAGES[@]}; do
	# Rewrite import paths
	rewrite_imports $package
done
