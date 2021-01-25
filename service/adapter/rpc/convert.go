package rpc

import (
	"github.com/golang/protobuf/proto"

	"github.com/xuperchain/xupercore/bcs/ledger/xledger/xldgpb"
	ecom "github.com/xuperchain/xupercore/kernel/engines/xuperos/commom"
	"github.com/xuperchain/xuperos/common/xupospb/pb"
)

// 错误映射配置
var StdErrToXchainErrMap = map[int]pb.XChainErrorEnum{
	ecom.ErrSuccess.Code:      pb.XChainErrorEnum_SUCCESS,
	ecom.ErrInternal.Code:     pb.XChainErrorEnum_UNKNOW_ERROR,
	ecom.ErrUnknown.Code:      pb.XChainErrorEnum_UNKNOW_ERROR,
	ecom.ErrForbidden.Code:    pb.XChainErrorEnum_CONNECT_REFUSE,
	ecom.ErrUnauthorized.Code: pb.XChainErrorEnum_CONNECT_REFUSE,
	ecom.ErrParameter.Code:    pb.XChainErrorEnum_CONNECT_REFUSE,
}

// 为了完全兼容老版本pb结构，转换交易结构
func TxToXledger(tx *pb.Transaction) *xldgpb.Transaction {
	if tx == nil {
		return nil
	}

	prtBuf, err := proto.Marshal(tx)
	if err != nil {
		return nil
	}

	var newTx xldgpb.Transaction
	err = proto.Unmarshal(prtBuf, &newTx)
	if err != nil {
		return nil
	}

	return &newTx
}

// 为了完全兼容老版本pb结构，转换交易结构
func TxToXchain(tx *xldgpb.Transaction) *pb.Transaction {
	if tx == nil {
		return nil
	}

	prtBuf, err := proto.Marshal(tx)
	if err != nil {
		return nil
	}

	var newTx pb.Transaction
	err = proto.Unmarshal(prtBuf, &newTx)
	if err != nil {
		return nil
	}

	return &newTx
}

// 为了完全兼容老版本pb结构，转换区块结构
func BlockToXledger(block *pb.InternalBlock) *xldgpb.InternalBlock {
	if block == nil {
		return nil
	}

	blkBuf, err := proto.Marshal(block)
	if err != nil {
		return nil
	}

	var newBlock xldgpb.InternalBlock
	err = proto.Unmarshal(blkBuf, &newBlock)
	if err != nil {
		return nil
	}

	return &newBlock
}

// 为了完全兼容老版本pb结构，转换区块结构
func BlockToXchain(block *xldgpb.InternalBlock) *pb.InternalBlock {
	if block == nil {
		return nil
	}

	blkBuf, err := proto.Marshal(block)
	if err != nil {
		return nil
	}

	var newBlock pb.InternalBlock
	err = proto.Unmarshal(blkBuf, &newBlock)
	if err != nil {
		return nil
	}

	return &newBlock
}
