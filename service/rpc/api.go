package rpc

import (
	"context"

	sctx "github.com/xuperchain/xupercore/example/xchain/common/context"
	pb "github.com/xuperchain/xuperos/common/xupospb"
)

// 注意：
// 1.rpc接口响应resp不能为nil，必须实例化
// 2.rpc接口响应err必须为ecom.Error类型的标准错误，没有错误响应err=nil
// 3.rpc接口不需要关注resp.Header，由拦截器根据err统一设置
// 4.rpc接口可以调用log库提供的SetInfoField方法附加输出到ending log

// 示例接口
func (t *RpcServ) CheckAlive(gctx context.Context, req *pb.BaseReq) (*pb.BaseResp, error) {
	// 默认响应
	resp := &pb.BaseResp{}
	// 获取请求上下文，对内传递rctx
	rctx := sctx.ValueReqCtx(gctx)

	rctx.GetLog().Debug("check alive succ")
	return resp, nil
}
