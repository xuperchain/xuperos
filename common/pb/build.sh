#!/bin/bash

protoc -I ./ --go_out=plugins=grpc:./ ./xuperos.proto

#protoc ./github.com/xuperchain/xupercore/protos/ledger.proto --go_out=plugins=grpc:.
#protoc ./github.com/xuperchain/xupercore/protos/contract.proto --go_out=plugins=grpc:.
#protoc ./github.com/xuperchain/xupercore/protos/permission.proto --go_out=plugins=grpc:.
#protoc ./github.com/xuperchain/xupercore/bcs/ledger/xledger/pb/xledger.proto --go_out=plugins=grpc:.

protoc -I ./github.com/ownluke/xuperos/common/pb/pb/googleapis \
       -I ./ ./github.com/ownluke/xuperos/common/pb/rpc.proto \
       --go_out=plugins=grpc:.
