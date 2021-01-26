package common

import (
	"fmt"

	"github.com/xuperchain/xuperos/common/xupospb/pb"
)

// 适配原结构计算txid
func MakeTxId(tx *pb.Transaction) ([]byte, error) {
	// 转化结构

	// 计算txid
}

// 适配原结构签名
func ComputeTxSign(tx *pb.Transaction) ([]byte, error) {
	// 转换结构

	// 计算txid
}
