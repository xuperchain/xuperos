#!/bin/bash

protoc -I ../ \
--go_opt=paths=source_relative \
--go_out=plugins=grpc:./ \
--grpc-gateway_out=logtostderr=true:./ \
-I common/pb/googleapis \
-I ./ common/pb/rpc.proto

protoc -I ../ \
--go_opt=paths=source_relative \
--go_out=plugins=grpc:./ \
--grpc-gateway_out=logtostderr=true:./ \
-I common/pb/googleapis \
-I ./ common/pb/xendorser.proto