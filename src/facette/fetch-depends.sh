#!/bin/bash

read -d '' DEPENDS <<EOF
git  https://github.com/fatih/set         master
git  https://github.com/gorilla/context   master
git  https://github.com/gorilla/handlers  master
git  https://github.com/gorilla/mux       master
git  https://github.com/nu7hatch/gouuid   master
git  https://github.com/ziutek/rrd        master
EOF

fetch_git() {
	NAME=$1
	URL=$2
	BRANCH=$3

	if [ ! -d "$SRC_DIR/$NAME" ]; then
		echo "Fetching $NAME..."
		mkdir -p $SRC_DIR/$NAME
		git clone $URL $SRC_DIR/$NAME -b $BRANCH
	elif [ $UPDATE -ne 0 ]; then
		echo "Updating $NAME..."
		git --git-dir $SRC_DIR/$NAME/.git pull
	fi
}

fetch_hg() {
	NAME=$1
	URL=$2
	BRANCH=$3

	if [ ! -d "$SRC_DIR/$NAME" ]; then
		echo "Fetching $NAME..."
		mkdir -p $SRC_DIR/$NAME
		hg clone $URL $SRC_DIR/$NAME -b $BRANCH
	elif [ $UPDATE -ne 0 ]; then
		echo "Updating $NAME..."
		hg -R $SRC_DIR/$NAME pull
		hg -R $SRC_DIR/$NAME update
	fi
}

print_usage() {
	echo "Usage: $(basename $0) [-u] DIR" >&2
	exit 1
}

# Parse command-line arguments
declare -i UPDATE=0

[ $# -lt 1 -o $# -gt 2 ] && print_usage

if [ $# -eq 1 ]; then
	SRC_DIR=$1
else
	[ "$1" != '-u' ] && print_usage

	UPDATE=1
	SRC_DIR=$2
fi

# Fetch dependencies
IFS=$'\n'

for ENTRY in $DEPENDS; do
	unset IFS
	read TYPE URL BRANCH <<<$(echo $ENTRY)

	fetch_$TYPE ${URL#http*://} $URL $BRANCH
done
