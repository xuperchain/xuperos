module github.com/xuperchain/xuperos

go 1.14

require (
	github.com/ChainSafe/go-schnorrkel v0.0.0-20200626160457-b38283118816 // indirect
	github.com/golang/protobuf v1.4.2
	github.com/google/gofuzz v1.1.1-0.20200604201612-c04b05f3adfa // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-gateway v1.15.2
	github.com/hyperledger/burrow v0.30.5
	github.com/manifoldco/promptui v0.7.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.6.2
	github.com/xuperchain/crypto v0.0.0-20201028025054-4d560674bcd6
	github.com/xuperchain/xupercore v0.0.0-20210524075332-3896dfb6780c
	golang.org/x/mod v0.1.1-0.20191209134235-331c550502dd // indirect
	golang.org/x/net v0.0.0-20200822124328-c89045814202
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013
	google.golang.org/grpc v1.32.0
	google.golang.org/protobuf v1.24.0 // indirect
)

replace github.com/hyperledger/burrow => github.com/xuperchain/burrow v0.30.6-0.20210115120720-3da1be35a1e2
