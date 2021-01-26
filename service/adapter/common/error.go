package common

// 错误映射配置
var StdErrToXchainErrMap = map[int]pb.XChainErrorEnum{
	ecom.ErrSuccess.Code:      pb.XChainErrorEnum_SUCCESS,
	ecom.ErrInternal.Code:     pb.XChainErrorEnum_UNKNOW_ERROR,
	ecom.ErrUnknown.Code:      pb.XChainErrorEnum_UNKNOW_ERROR,
	ecom.ErrForbidden.Code:    pb.XChainErrorEnum_CONNECT_REFUSE,
	ecom.ErrUnauthorized.Code: pb.XChainErrorEnum_CONNECT_REFUSE,
	ecom.ErrParameter.Code:    pb.XChainErrorEnum_CONNECT_REFUSE,
}
