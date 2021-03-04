#!/bin/bash

cd `dirname $0`/../

HOMEDIR=`pwd`
OUTDIR="$HOMEDIR/xvm"
XVMPKG="https://github.com/xuperchain/xvm/archive/main.zip"

# make output dir
rm -rf $OUTDIR
mkdir -p $OUTDIR

function buildxvm() {
    wget -O "$OUTDIR/xvm.zip" "$XVMPKG" --no-check-certificate
    if [ $? != 0 ]; then
        echo "download xvm failed"
        exit 1
    fi

    unzip -d "$OUTDIR" "$OUTDIR/xvm.zip"
    mv "$OUTDIR/xvm-main" "$OUTDIR/xvm"

    make -C "$OUTDIR/xvm/compile/wabt" -j 4
    if [ $? != 0 ]; then
        echo "complie xvm failed"
        exit 1
    fi

    cp -r "$OUTDIR/xvm/compile/wabt/build/wasm2c" "$OUTDIR"
}

# build xvm
if [ ! -d "$OUTDIR/wasm2c" ]; then
    buildxvm
fi

echo "compile done!"
