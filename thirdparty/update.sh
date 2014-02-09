#!/bin/bash

read -d '' DEPENDS <<EOF
git  https://github.com/etix/stoppableListener  master
git  https://github.com/fatih/set               master
git  https://github.com/gorilla/context         master
git  https://github.com/gorilla/handlers        master
git  https://github.com/gorilla/mux             master
git  https://github.com/howeyc/gopass           master
git  https://github.com/nu7hatch/gouuid         master
git  https://github.com/ziutek/rrd              master
hg   https://code.google.com/p/go.crypto        default
EOF

fetch_git() {
	NAME=$1
	URL=$2
	BRANCH=$3

	if [ ! -d "$SRC_DIR/checkouts/$NAME" ]; then
		echo "Fetching $NAME..."
		mkdir -p $SRC_DIR/checkouts/$NAME
		git clone --quiet $URL $SRC_DIR/checkouts/$NAME -b $BRANCH
	else
		echo "Updating $NAME..."
		git -C $SRC_DIR/checkouts/$NAME pull --quiet
	fi
}

fetch_hg() {
	NAME=$1
	URL=$2
	BRANCH=$3

	if [ ! -d "$SRC_DIR/checkouts/$NAME" ]; then
		echo "Fetching $NAME..."
		mkdir -p $SRC_DIR/checkouts/$NAME
		hg clone --quiet $URL $SRC_DIR/checkouts/$NAME -b $BRANCH
	else
		echo "Updating $NAME..."
		hg -R $SRC_DIR/checkouts/$NAME pull --quiet
		hg -R $SRC_DIR/checkouts/$NAME update --quiet
	fi
}

rewrite_imports() {
	NAME=$1

	echo "Rewriting $NAME import paths..."
	find $SRC_DIR -type f ! -path "$SRC_DIR/checkouts/*" -name '*.go' \
		-exec sed -e "s@\"$NAME@\"github.com/facette/facette/thirdparty/$NAME@" -i {} \;
}

print_usage() {
	echo "Usage: $(basename $0) [-u]" >&2
	exit 1
}

# Parse command-line arguments
[ $# -ne 0 ] && print_usage

SRC_DIR=$(dirname $0)

# Fetch dependencies
declare -a PACKAGES

IFS=$'\n'; for ENTRY in $DEPENDS; do
	unset IFS
	read TYPE URL BRANCH <<<$(echo $ENTRY)

	# Get package name
	NAME=${URL#http*://}

	# Fetch project source and create a local copy
	fetch_$TYPE $NAME $URL $BRANCH

	rm -rf $SRC_DIR/$NAME
	mkdir -p $SRC_DIR/$NAME
	cp -a $SRC_DIR/checkouts/$NAME/* $SRC_DIR/$NAME

	PACKAGES+=($NAME)
done

for NAME in ${PACKAGES[@]}; do
	# Rewrite import paths
	rewrite_imports $NAME
done
