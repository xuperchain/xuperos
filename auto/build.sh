#!/bin/bash

cd `dirname $0`/../

HOMEDIR=`pwd`
OUTDIR="$HOMEDIR/output"

# make output dir
if [ ! -d "$OUTDIR" ];then
    mkdir $OUTDIR
fi
rm -rf "$OUTDIR/*"

function buildpkg() {
    output=$1
    pkg=$2

    version=`git rev-parse --abbrev-ref HEAD`
    commitId=`git rev-parse --short HEAD`
    buildTime=$(date "+%Y-%m-%d-%H:%M:%S")
    
    ldflags="-X version.Version=$version -X version.BuildTime=$buildTime -X version.CommitID=$commitId"
    
    # build
    if [ ! -d "$OUTDIR/bin" ];then
        mkdir "$OUTDIR/bin"
    fi
    go build -o "$OUTDIR/bin/$output" -ldflags $ldflags $pkg
}

# build xuperos
buildpkg xuperos "$HOMEDIR/cmd/xuperos/main.go"
buildpkg xuperos-cli "$HOMEDIR/cmd/client/main.go"

# build output
cp -r "$HOMEDIR/conf" "$OUTDIR"
cp "$HOMEDIR/auto/control.sh" "$OUTDIR"
mkdir "$OUTDIR/data" 
