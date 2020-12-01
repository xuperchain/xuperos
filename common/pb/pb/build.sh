#!/bin/bash

protoc -I ./ ./*.proto \
    -I ./googleapis \
    --go_out=plugins=grpc:./ \
    --grpc-gateway_out=logtostderr=true:./
