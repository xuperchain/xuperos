package common

import (
	ecom "github.com/xuperchain/xupercore/kernel/engines/xuperos/common"
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
