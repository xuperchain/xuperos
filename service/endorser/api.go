package endorser

import (
	"context"
	ecom "github.com/xuperchain/xupercore/kernel/engines/xuperos/common"
	"github.com/xuperchain/xupercore/lib/utils"

	sctx "github.com/xuperchain/xupercore/example/xchain/common/context"
	"github.com/xuperchain/xuperos/common/xupospb/pb"
)

func (t *RpcServ) EndorserCall(gctx context.Context, req *pb.EndorserRequest) (*pb.EndorserResponse, error) {
	resp := &pb.EndorserResponse{}
	rctx := sctx.ValueReqCtx(gctx)
	// 校验参数
	if req == nil || req.GetFee() == nil || req.GetRequestName() == "" || req.GetBcName() == "" {
		rctx.GetLog().Warn("param error,some param unset")
		return resp, ecom.ErrParameter
	}
	endc, err := t.getClient(t.getHost())
	if err != nil {
		return nil, err
	}
	res, err := endc.EndorserCall(gctx, req)
	if err != nil {
		return nil, err
	}
	resp.EndorserAddress = res.EndorserAddress
	resp.ResponseName = res.ResponseName
	resp.ResponseData = res.ResponseData
	resp.EndorserSign = res.EndorserSign
	rctx.GetLog().SetInfoField("bc_name", req.GetBcName())
	rctx.GetLog().SetInfoField("request_name", req.GetBcName())
	rctx.GetLog().SetInfoField("txid", utils.F(req.GetFee().Txid))
	return resp, nil
}
