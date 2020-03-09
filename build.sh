#!/bin/bash

# get xuperchain
update() {
    echo 'update dependency'
    mkdir -p deps
    cd deps
    git clone https://github.com/xuperchain/xuperchain.git
    cd xuperchain && git checkout v3.5.0
    cd ../../
}

# build dependencies and xuperos
build() {
    # build xuperchain
    mkdir -p output
    cd ./deps/xuperchain
    make
    cd ../../
    cp -r ./deps/xuperchain/output/* ./output

    ## build xuperos
    cp -r ./conf/* ./output/conf/
    cp ./data/config/xuper.json  ./output/data/config/xuper.json
}

clean() {
    rm -rf ./output
    rm -rf ./deps
}

case C"$1" in
    C)
        update
        build
        echo 'Done!'
        ;;
    Cupdate)
        update
        echo 'Done!'
        ;;
    Cbuild)
        build
        echo 'Done!'
        ;;
    Cclean)
       clean
       echo 'Done!'
       ;;
esac