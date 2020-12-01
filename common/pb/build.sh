#!/bin/bash

protoc -I ./ --go_out=plugins=grpc:./ ./xuperos.proto
