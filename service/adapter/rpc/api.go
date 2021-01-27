package rpc

import (
	"context"
	sctx "github.com/xuperchain/xupercore/example/xchain/common/context"
	ecom "github.com/xuperchain/xupercore/kernel/engines/xuperos/common"
	"github.com/xuperchain/xupercore/lib/utils"
	"math/big"

	"github.com/xuperchain/xupercore/bcs/ledger/xledger/xldgpb"
	"github.com/xuperchain/xupercore/kernel/engines/xuperos/reader"
	"github.com/xuperchain/xupercore/kernel/network/p2p"
	"github.com/xuperchain/xupercore/protos"
	"github.com/xuperchain/xuperos/models"
	acom "github.com/xuperchain/xuperos/service/adapter/common"

	rctx "github.com/xuperchain/xuperos/common/context"
	"github.com/xuperchain/xuperos/common/xupospb/pb"
)

// 注意：
// 1.rpc接口响应resp不能为nil，必须实例化
// 2.rpc接口响应err必须为ecom.Error类型的标准错误，没有错误响应err=nil
// 3.rpc接口不需要关注resp.Header，由拦截器根据err统一设置
// 4.rpc接口可以调用log库提供的SetInfoField方法附加输出到ending log

// PostTx post transaction to blockchain network
func (t *RpcServ) PostTx(gctx context.Context, req *pb.TxStatus) (*pb.CommonReply, error) {
	// 默认响应
	resp := &pb.CommonReply{}
	// 获取请求上下文，对内传递rctx
	rctx := sctx.ValueReqCtx(gctx)

	// 校验参数
	if req == nil || req.GetTx() == nil || req.GetBcname() == "" {
		rctx.GetLog().Warn("param error,some param unset")
		return resp, ecom.ErrParameter
	}
	tx := acom.TxToXledger(req.GetTx())
	if tx == nil {
		rctx.GetLog().Warn("param error,tx convert to xledger tx failed")
		return resp, ecom.ErrParameter
	}

	// 提交交易
	handle, err := models.NewChainHandle(req.GetBcname(), rctx)
	if err != nil {
		rctx.GetLog().Warn("new chain handle failed", "err", err.Error())
		return resp, err
	}
	err = handle.SubmitTx(tx)
	rctx.GetLog().SetInfoField("bc_name", req.GetBcname())
	rctx.GetLog().SetInfoField("txid", utils.F(req.GetTxid()))
	return resp, err
}

// PreExec smart contract preExec process
func (t *RpcServ) PreExec(gctx context.Context, req *pb.InvokeRPCRequest) (*pb.InvokeRPCResponse, error) {
	// 默认响应
	resp := &pb.InvokeRPCResponse{}
	// 获取请求上下文，对内传递rctx
	rctx := sctx.ValueReqCtx(gctx)

	// 校验参数
	if req == nil || req.GetBcname() == "" || len(req.GetRequests()) < 1 {
		rctx.GetLog().Warn("param error,some param unset")
		return resp, ecom.ErrParameter
	}
	reqs, err := acom.ConvertInvokeReq(req.GetRequests())
	if err != nil {
		rctx.GetLog().Warn("param error, convert failed", "err", err)
		return resp, ecom.ErrParameter
	}

	// 预执行
	handle, err := models.NewChainHandle(req.GetBcname(), rctx)
	if err != nil {
		rctx.GetLog().Warn("new chain handle failed", "err", err.Error())
		return resp, err
	}
	res, err := handle.PreExec(reqs, req.GetInitiator(), req.GetAuthRequire())
	rctx.GetLog().SetInfoField("bc_name", req.GetBcname())
	rctx.GetLog().SetInfoField("initiator", req.GetInitiator())
	// 设置响应
	if err == nil {
		resp.Bcname = req.GetBcname()
		resp.Response = acom.ConvertInvokeResp(res)
	}

	return resp, err
}

// PreExecWithSelectUTXO preExec + selectUtxo
func (t *RpcServ) PreExecWithSelectUTXO(gctx context.Context,
	req *pb.PreExecWithSelectUTXORequest) (*pb.PreExecWithSelectUTXOResponse, error) {

	// 默认响应
	resp := &pb.PreExecWithSelectUTXOResponse{}
	// 获取请求上下文，对内传递rctx
	rctx := sctx.ValueReqCtx(gctx)

	if req == nil || req.GetBcname() == "" || len(req.GetRequest()) < 1 {
		rctx.GetLog().Warn("param error,some param unset")
		return resp, ecom.ErrParameter
	}

	// PreExec
	preExecRes, err := t.PreExec(gctx, req.GetRequest())
	if err != nil {
		rctx.GetLog().Warn("pre exec failed", "err", err)
		return resp, err
	}

	// SelectUTXO
	totalAmount := req.GetTotalAmount() + preExecRes.GetResponse().GetGasUsed()
	if totalAmount < 1 {
		return resp, nil
	}
	utxoInput := &pb.UtxoInput{
		Header:    req.GetHeader(),
		Bcname:    req.GetBcname(),
		Address:   req.GetAddress(),
		Publickey: req.GetSignInfo().GetPublicKey(),
		TotalNeed: big.NewInt(totalAmount).String(),
		UserSign:  req.GetSignInfo().GetSign(),
		NeedLock:  req.GetNeedLock(),
	}
	utxoOut, err := t.SelectUTXO(gctx, utxoInput)
	if err != nil {
		return resp, err
	}
	utxoOut.Header = req.GetHeader()

	// 设置响应
	resp.Bcname = req.GetBcname()
	resp.Response = preExecRes.GetResponse()
	resp.UtxoOutput = utxoOut

	return resp, nil
}

// SelectUTXO select utxo inputs depending on amount
func (t *RpcServ) SelectUTXO(gctx context.Context, req *pb.UtxoInput) (*pb.UtxoOutput, error) {
	// 默认响应
	resp := &pb.UtxoOutput{}
	// 获取请求上下文，对内传递rctx
	rctx := sctx.ValueReqCtx(gctx)

	if req == nil || req.GetBcname() == "" || req.GetTotalNeed() == "" {
		rctx.GetLog().Warn("param error,some param unset")
		return resp, ecom.ErrParameter
	}
	totalNeed, ok := new(big.Int).SetString(req.GetTotalNeed(), 10)
	if !ok {
		rctx.GetLog().Warn("param error,total need set error", "totalNeed", req.GetTotalNeed())
		return resp, ecom.ErrParameter
	}

	// select utxo
	handle, err := models.NewChainHandle(req.GetBcname(), rctx)
	if err != nil {
		rctx.GetLog().Warn("new chain handle failed", "err", err.Error())
		return resp, err
	}
	out, err := handle.SelectUtxo(req.GetAddress(), totalNeed, req.GetNeedLock(), false,
		req.GetPublickey(), req.GetUserSign())
	if err != nil {
		rctx.GetLog().Warn("select utxo failed", "err", err.Error())
		return resp, err
	}

	utxoList, err := acom.UtxoListToXchain(out.GetUtxoList())
	if err != nil {
		rctx.GetLog().Warn("convert utxo failed", "err", err)
		return resp, ecom.ErrInternal
	}

	resp.UtxoList = utxoList
	resp.TotalSelected = out.GetTotalSelected()
	return resp, nil
}

// SelectUTXOBySize select utxo inputs depending on size
func (t *RpcServ) SelectUTXOBySize(gctx context.Context, req *pb.UtxoInput) (*pb.UtxoOutput, error) {
	// 默认响应
	resp := &pb.UtxoOutput{}
	// 获取请求上下文，对内传递rctx
	rctx := sctx.ValueReqCtx(gctx)

	if req == nil || req.GetBcname() == "" {
		rctx.GetLog().Warn("param error,some param unset")
		return resp, ecom.ErrParameter
	}

	// select utxo
	handle, err := models.NewChainHandle(req.GetBcname(), rctx)
	if err != nil {
		rctx.GetLog().Warn("new chain handle failed", "err", err.Error())
		return resp, err
	}
	out, err := handle.SelectUTXOBySize(req.GetAddress(), req.GetNeedLock(), false,
		req.GetPublickey(), req.GetUserSign())
	if err != nil {
		rctx.GetLog().Warn("select utxo failed", "err", err.Error())
		return resp, err
	}

	utxoList, err := acom.UtxoListToXchain(out.GetUtxoList())
	if err != nil {
		rctx.GetLog().Warn("convert utxo failed", "err", err)
		return resp, ecom.ErrInternal
	}

	resp.UtxoList = utxoList
	resp.TotalSelected = out.GetTotalSelected()
	return resp, nil
}

// QueryContractStatData query statistic info about contract
func (t *RpcServ) QueryContractStatData(gctx context.Context,
	req *pb.ContractStatDataRequest) (*pb.ContractStatDataResponse, error) {
	// 默认响应
	resp := &pb.ContractStatDataResponse{}
	// 获取请求上下文，对内传递rctx
	rctx := sctx.ValueReqCtx(gctx)

	if req == nil || req.GetBcname() == "" {
		rctx.GetLog().Warn("param error,some param unset")
		return resp, ecom.ErrParameter
	}

	handle, err := models.NewChainHandle(req.GetBcname(), rctx)
	if err != nil {
		rctx.GetLog().Warn("new chain handle failed", "err", err.Error())
		return resp, err
	}
	res, err := handle.QueryContractStatData()
	if err != nil {
		rctx.GetLog().Warn("query contract stat data failed", "err", err.Error())
		return resp, err
	}

	resp.Bcname = req.GetBcname()
	resp.Data = &pb.ContractStatData{
		AccountCount:  res.GetAccountCount(),
		ContractCount: res.GetContractCount(),
	}

	return resp, nil
}

// QueryUtxoRecord query utxo records
func (t *RpcServ) QueryUtxoRecord(gctx context.Context,
	req *pb.UtxoRecordDetail) (*pb.UtxoRecordDetail, error) {

	// 默认响应
	resp := &pb.UtxoRecordDetail{}
	// 获取请求上下文，对内传递rctx
	rctx := sctx.ValueReqCtx(gctx)

	if req == nil || req.GetBcname() == "" {
		rctx.GetLog().Warn("param error,some param unset")
		return resp, ecom.ErrParameter
	}

	handle, err := models.NewChainHandle(req.GetBcname(), rctx)
	if err != nil {
		rctx.GetLog().Warn("new chain handle failed", "err", err.Error())
		return resp, err
	}
	res, err := handle.QueryUtxoRecord(req.GetAccountName(), req.GetDisplayCount())
	if err != nil {
		rctx.GetLog().Warn("query utxo record failed", "err", err.Error())
		return resp, err
	}

	resp.Bcname = req.GetBcname()
	resp.AccountName = req.GetAccountName()
	resp.OpenUtxoRecord = acom.UtxoRecordToXchain(res.GetOpenUtxo())
	resp.LockedUtxoRecord = acom.UtxoRecordToXchain(res.GetLockedUtxo())
	resp.FrozenUtxoRecord = acom.UtxoRecordToXchain(res.GetFrozenUtxo())
	resp.DisplayCount = req.GetDisplayCount()

	return resp, nil
}

// QueryACL query some account info
func (t *RpcServ) QueryACL(gctx context.Context, req *pb.AclStatus) (*pb.AclStatus, error) {
	// 默认响应
	resp := &pb.AclStatus{}
	// 获取请求上下文，对内传递rctx
	rctx := sctx.ValueReqCtx(gctx)

	if req == nil || req.GetBcname() == "" {
		rctx.GetLog().Warn("param error,some param unset")
		return resp, ecom.ErrParameter
	}
	if len(req.GetAccountName()) < 1 && (len(req.GetContractName()) < 1 || len(req.GetMethodName()) < 1) {
		rctx.GetLog().Warn("param error,unset name")
		return resp, ecom.ErrParameter
	}

	handle, err := models.NewChainHandle(req.GetBcname(), rctx)
	if err != nil {
		rctx.GetLog().Warn("new chain handle failed", "err", err.Error())
		return resp, err
	}

	var aclRes *protos.Acl
	if len(req.GetAccountName()) > 0 {
		aclRes, err = handle.QueryAccountACL(req.GetAccountName())
	} else if len(req.GetContractName()) > 0 && len(req.GetMethodName()) > 0 {
		aclRes, err = handle.QueryContractMethodACL(req.GetContractName(), req.GetMethodName())
	}
	if err != nil {
		rctx.GetLog().Warn("query acl failed", "err", err)
		return resp, err
	}
	xchainAcl := acom.AclToXchain(aclRes)
	if xchainAcl == nil {
		rctx.GetLog().Warn("convert acl failed")
		return resp, ecom.ErrInternal
	}

	resp.AccountName = req.GetAccountName()
	resp.ContractName = req.GetContractName()
	resp.MethodName = req.GetMethodName()
	resp.Confirmed = true
	resp.Acl = xchainAcl

	return resp, nil
}

// GetAccountContracts get account request
func (s *RpcServ) GetAccountContracts(gctx context.Context, req *pb.GetAccountContractsRequest) (*pb.GetAccountContractsResponse, error) {
	// 默认响应
	resp := &pb.GetAccountContractsResponse{}
	// 获取请求上下文，对内传递rctx
	rctx := sctx.ValueReqCtx(gctx)

	if req == nil || req.GetBcname() == "" || req.GetAccount() == "" {
		rctx.GetLog().Warn("param error,some param unset")
		return resp, ecom.ErrParameter
	}

	handle, err := models.NewChainHandle(req.GetBcname(), rctx)
	if err != nil {
		rctx.GetLog().Warn("new chain handle failed", "err", err.Error())
		return resp, err
	}

	var res []*protos.ContractStatus
	res, err = handle.GetAccountContracts(req.GetAccount())
	if err != nil {
		rctx.GetLog().Warn("get account contract failed", "err", err)
		return resp, err
	}
	xchainContractStatus, err := acom.ContractStatusListToXchain(res)
	if xchainContractStatus == nil {
		rctx.GetLog().Warn("convert acl failed")
		return resp, ecom.ErrInternal
	}

	resp.ContractsStatus = xchainContractStatus

	rctx.GetLog().SetInfoField("bc_name", req.GetBcname())
	rctx.GetLog().SetInfoField("account", req.GetAccount())
	return resp, nil
}

// QueryTx Get transaction details
func (s *RpcServ) QueryTx(gctx context.Context, req *pb.TxStatus) (*pb.TxStatus, error) {
	// 默认响应
	resp := &pb.TxStatus{}
	// 获取请求上下文，对内传递rctx
	rctx := sctx.ValueReqCtx(gctx)

	if req == nil || req.GetBcname() == "" || len(req.GetTxid()) == 0 {
		rctx.GetLog().Warn("param error,some param unset")
		return resp, ecom.ErrParameter
	}

	handle, err := models.NewChainHandle(req.GetBcname(), rctx)
	if err != nil {
		rctx.GetLog().Warn("new chain handle failed", "err", err.Error())
		return resp, err
	}

	txInfo, err := handle.QueryTx(req.GetTxid())
	if err != nil {
		rctx.GetLog().Warn("query tx failed", "err", err)
		return resp, err
	}

	tx := acom.TxToXchain(txInfo.Tx)
	if tx == nil {
		rctx.GetLog().Warn("convert tx failed")
		return resp, ecom.ErrInternal
	}
	resp.Bcname = req.GetBcname()
	resp.Txid = req.GetTxid()
	resp.Tx = tx
	resp.Status = pb.TransactionStatus(txInfo.Status)
	resp.Distance = txInfo.Distance

	rctx.GetLog().SetInfoField("bc_name", req.GetBcname())
	rctx.GetLog().SetInfoField("account", utils.F(req.GetTxid()))
	return out, nil
}

// GetBalance get balance for account or addr
func (s *RpcServ) GetBalance(gctx context.Context, req *pb.AddressStatus) (*pb.AddressStatus, error) {
	// 默认响应
	resp := &pb.AddressStatus{}
	// 获取请求上下文，对内传递rctx
	rctx := sctx.ValueReqCtx(gctx)

	if req == nil || req.GetAddress() == "" {
		rctx.GetLog().Warn("param error,some param unset")
		return resp, ecom.ErrParameter
	}

	for i := 0; i < len(req.Bcs); i++ {
		handle, err := models.NewChainHandle(req.Bcs[i].Bcname, rctx)
		if err != nil {
			resp.Bcs[i].Error = pb.XChainErrorEnum_BLOCKCHAIN_NOTEXIST
			resp.Bcs[i].Balance = ""
			continue
		}
		balance, err := handle.GetBalance(req.Address)
		if err != nil {
			resp.Bcs[i].Error = pb.XChainErrorEnum_UNKNOW_ERROR
			resp.Bcs[i].Balance = ""
		} else {
			resp.Bcs[i].Error = pb.XChainErrorEnum_SUCCESS
			resp.Bcs[i].Balance = balance
		}
	}
	resp.Address = req.GetAddress()

	rctx.GetLog().SetInfoField("account", req.GetAddress())
	return resp, nil
}

// GetFrozenBalance get balance frozened for account or addr
func (s *RpcServ) GetFrozenBalance(gctx context.Context, req *pb.AddressStatus) (*pb.AddressStatus, error) {
	// 默认响应
	resp := &pb.AddressStatus{}
	// 获取请求上下文，对内传递rctx
	rctx := sctx.ValueReqCtx(gctx)

	if req == nil || req.GetAddress() == "" {
		rctx.GetLog().Warn("param error,some param unset")
		return resp, ecom.ErrParameter
	}

	for i := 0; i < len(req.Bcs); i++ {
		handle, err := models.NewChainHandle(req.Bcs[i].Bcname, rctx)
		if err != nil {
			resp.Bcs[i].Error = pb.XChainErrorEnum_BLOCKCHAIN_NOTEXIST
			resp.Bcs[i].Balance = ""
			continue
		}
		balance, err := handle.GetFrozenBalance(req.Address)
		if err != nil {
			resp.Bcs[i].Error = pb.XChainErrorEnum_UNKNOW_ERROR
			resp.Bcs[i].Balance = ""
		} else {
			resp.Bcs[i].Error = pb.XChainErrorEnum_SUCCESS
			resp.Bcs[i].Balance = balance
		}
	}
	resp.Address = req.GetAddress()

	rctx.GetLog().SetInfoField("account", req.GetAddress())
	return resp, nil
}

// GetBalanceDetail get balance frozened for account or addr
func (s *RpcServ) GetBalanceDetail(gctx context.Context, req *pb.AddressBalanceStatus) (*pb.AddressBalanceStatus, error) {
	// 默认响应
	resp := &pb.AddressBalanceStatus{}
	// 获取请求上下文，对内传递rctx
	rctx := sctx.ValueReqCtx(gctx)

	if req == nil || req.GetAddress() == "" {
		rctx.GetLog().Warn("param error,some param unset")
		return resp, ecom.ErrParameter
	}

	for i := 0; i < len(req.Tfds); i++ {
		handle, err := models.NewChainHandle(req.Tfds[i].Bcname, rctx)
		if err != nil {
			resp.Tfds[i].Error = pb.XChainErrorEnum_BLOCKCHAIN_NOTEXIST
			resp.Tfds[i].Tfd = nil
		}
		tfd, err := handle.GetBalanceDetail(req.GetAddress())
		if err != nil {
			resp.Tfds[i].Error = pb.XChainErrorEnum_UNKNOW_ERROR
			resp.Tfds[i].Tfd = nil
		} else {
			xchainTfd, err := acom.BalanceDetailsToXchain(tfd)
			if err != nil {
				resp.Tfds[i].Error = pb.XChainErrorEnum_UNKNOW_ERROR
				resp.Tfds[i].Tfd = nil
			}
			resp.Tfds[i].Error = pb.XChainErrorEnum_SUCCESS
			resp.Tfds[i].Tfd = xchainTfd
		}
	}

	rctx.GetLog().SetInfoField("account", req.GetAddress())
	return in, nil
}

// GetBlock get block info according to blockID
func (s *RpcServ) GetBlock(gctx context.Context, req *pb.BlockID) (*pb.Block, error) {
	// 默认响应
	resp := &pb.Block{}
	// 获取请求上下文，对内传递rctx
	rctx := sctx.ValueReqCtx(gctx)

	if req == nil || req.GetBcname() == "" || len(req.GetBlockid()) == 0 {
		rctx.GetLog().Warn("param error,some param unset")
		return resp, ecom.ErrParameter
	}

	handle, err := models.NewChainHandle(req.GetBcname(), rctx)
	if err != nil {
		rctx.GetLog().Warn("new chain handle failed", "err", err.Error())
		return resp, err
	}

	blockInfo, err := handle.QueryBlock(req.GetBlockid(), true)
	if err != nil {
		rctx.GetLog().Warn("query block error", "error", err)
		return resp, err
	}

	block := acom.BlockToXchain(blockInfo.Block)
	if err == nil {
		rctx.GetLog().Warn("convert block failed")
		return resp, ecom.ErrInternal
	}
	resp.Block = block
	resp.Status = pb.Block_EBlockStatus(blockInfo.Status)
	resp.Bcname = req.Bcname
	resp.Blockid = req.Blockid

	rctx.GetLog().SetInfoField("blockid", req.GetBlockid())
	rctx.GetLog().SetInfoField("height", blockInfo.GetBlock().GetHeight())
	return resp, nil
}

// GetBlockChainStatus get systemstatus
func (s *RpcServ) GetBlockChainStatus(gctx context.Context, req *pb.BCStatus) (*pb.BCStatus, error) {
	// 默认响应
	resp := &pb.BCStatus{}
	// 获取请求上下文，对内传递rctx
	rctx := sctx.ValueReqCtx(gctx)

	if req == nil || req.GetBcname() == "" {
		rctx.GetLog().Warn("param error,some param unset")
		return resp, ecom.ErrParameter
	}

	handle, err := models.NewChainHandle(req.GetBcname(), rctx)
	if err != nil {
		rctx.GetLog().Warn("new chain handle failed", "err", err.Error())
		return resp, err
	}

	status, err := handle.QueryChainStatus()
	if err != nil {
		rctx.GetLog().Warn("get chain status error", "error", err)
		return resp, err
	}

	block := acom.BlockToXchain(status.Block)
	if block == nil {
		rctx.GetLog().Warn("convert block failed")
		return resp, err
	}
	ledgerMeta := acom.LedgerMetaToXchain(status.LedgerMeta)
	if ledgerMeta == nil {
		rctx.GetLog().Warn("convert ledger meta failed")
		return resp, err
	}
	utxoMeta := acom.UtxoMetaToXchain(status.UtxoMeta)
	if utxoMeta == nil {
		rctx.GetLog().Warn("convert utxo meta failed")
		return resp, err
	}
	resp.Meta = ledgerMeta
	resp.Block = block
	resp.UtxoMeta = utxoMeta

	rctx.GetLog().SetInfoField("bc_name", req.GetBcname())
	rctx.GetLog().SetInfoField("blockid", utils.F(resp.Block.Blockid))
	return resp, nil
}

// ConfirmBlockChainStatus confirm is_trunk
func (s *RpcServ) ConfirmBlockChainStatus(ctx context.Context, in *pb.BCStatus) (*pb.BCTipStatus, error) {
	reqCtx := rctx.ReqCtxFromContext(ctx)
	out := &pb.BCTipStatus{Header: defRespHeader(in.Header)}

	chain, err := s.engine.Get(in.GetBcname())
	if err != nil {
		out.Header.Error = pb.XChainErrorEnum_BLOCKCHAIN_NOTEXIST
		reqCtx.GetLog().Warn("block chain not exists", "bc", in.GetBcname())
		return out, err
	}

	chainReader := reader.NewChainReader(chain.Context(), reqCtx)
	isTrunkTip, err := chainReader.IsTrunkTipBlock(in.GetBlock().GetBlockid())
	if err != nil {
		return nil, err
	}

	out.IsTrunkTip = isTrunkTip
	return out, nil
}

// GetBlockChains get BlockChains
func (s *RpcServ) GetBlockChains(ctx context.Context, in *pb.CommonIn) (*pb.BlockChains, error) {
	out := &pb.BlockChains{Header: defRespHeader(in.Header)}
	out.Blockchains = s.engine.GetChains()
	return out, nil
}

// GetSystemStatus get systemstatus
func (s *RpcServ) GetSystemStatus(ctx context.Context, in *pb.CommonIn) (*pb.SystemsStatusReply, error) {
	reqCtx := rctx.ReqCtxFromContext(ctx)
	out := &pb.SystemsStatusReply{Header: defRespHeader(in.Header)}

	systemsStatus := &pb.SystemsStatus{
		Header: in.Header,
		Speeds: &pb.Speeds{
			SumSpeeds: make(map[string]float64),
			BcSpeeds:  make(map[string]*pb.BCSpeeds),
		},
	}
	bcs := s.engine.GetChains()
	for _, bcName := range bcs {
		bcStatus := &pb.BCStatus{Header: in.Header, Bcname: bcName}
		status, err := s.GetBlockChainStatus(ctx, bcStatus)
		if err != nil {
			reqCtx.GetLog().Warn("get chain status error", "error", err)
		}

		systemsStatus.BcsStatus = append(systemsStatus.BcsStatus, status)
	}

	if in.ViewOption == pb.ViewOption_NONE || in.ViewOption == pb.ViewOption_PEERS {
		peerInfo := s.engine.Context().Net.PeerInfo()
		peerUrls := make([]string, 0, len(peerInfo.Peer))
		for _, peer := range peerInfo.Peer {
			peerUrls = append(peerUrls, peer.Address)
		}
		systemsStatus.PeerUrls = peerUrls
	}

	out.SystemsStatus = systemsStatus
	return out, nil
}

// GetNetURL get net url in p2p_base
func (s *RpcServ) GetNetURL(ctx context.Context, in *pb.CommonIn) (*pb.RawUrl, error) {
	out := &pb.RawUrl{Header: defRespHeader(in.Header)}
	peerInfo := s.engine.Context().Net.PeerInfo()
	out.RawUrl = peerInfo.Address
	return out, nil
}

// GetBlockByHeight  get trunk block by height
func (s *RpcServ) GetBlockByHeight(ctx context.Context, in *pb.BlockHeight) (*pb.Block, error) {
	reqCtx := rctx.ReqCtxFromContext(ctx)
	out := &pb.Block{Header: defRespHeader(in.Header)}

	chain, err := s.engine.Get(in.GetBcname())
	if err != nil {
		out.Header.Error = pb.XChainErrorEnum_BLOCKCHAIN_NOTEXIST
		reqCtx.GetLog().Warn("block chain not exists", "bc", in.GetBcname())
		return out, err
	}

	ledgerReader := reader.NewLedgerReader(chain.Context(), reqCtx)
	blockInfo, err := ledgerReader.QueryBlockByHeight(in.Height, true)
	if err != nil {
		out.Header.Error = pb.XChainErrorEnum_BLOCK_EXIST_ERROR
		reqCtx.GetLog().Warn("query block error", "bc", in.GetBcname(), "height", in.Height)
		return out, err
	}

	out.Block = blockInfo.Block
	out.Status = pb.Block_EBlockStatus(blockInfo.Status)

	transactions := out.GetBlock().GetTransactions()
	if transactions != nil {
		out.Block.Transactions = transactions
	}

	reqCtx.GetLog().SetInfoField("height", in.Height)
	reqCtx.GetLog().SetInfoField("blockid", out.GetBlockid())
	return out, nil
}

// GetAccountByAK get account list with contain ak
func (s *RpcServ) GetAccountByAK(ctx context.Context, in *pb.AK2AccountRequest) (*pb.AK2AccountResponse, error) {
	reqCtx := rctx.ReqCtxFromContext(ctx)
	out := &pb.AK2AccountResponse{Header: defRespHeader(in.Header), Bcname: in.GetBcname()}

	chain, err := s.engine.Get(in.GetBcname())
	if err != nil {
		out.Header.Error = pb.XChainErrorEnum_BLOCKCHAIN_NOTEXIST
		reqCtx.GetLog().Warn("block chain not exists", "bc", in.GetBcname())
		return out, err
	}

	contractReader := reader.NewContractReader(chain.Context(), reqCtx)
	accounts, err := contractReader.GetAccountByAK(in.GetAddress())
	if err != nil || accounts == nil {
		reqCtx.GetLog().Warn("QueryAccountContainAK error", "logid", out.Header.Logid, "error", err)
		return out, err
	}

	out.Account = accounts
	return out, err
}

// GetAddressContracts get contracts of accounts contain a specific address
func (s *RpcServ) GetAddressContracts(ctx context.Context, in *pb.AddressContractsRequest) (*pb.AddressContractsResponse, error) {
	reqCtx := rctx.ReqCtxFromContext(ctx)
	out := &pb.AddressContractsResponse{Header: defRespHeader(in.Header)}

	chain, err := s.engine.Get(in.GetBcname())
	if err != nil {
		out.Header.Error = pb.XChainErrorEnum_BLOCKCHAIN_NOTEXIST
		reqCtx.GetLog().Warn("block chain not exists", "bc", in.GetBcname())
		return out, err
	}

	contractReader := reader.NewContractReader(chain.Context(), reqCtx)
	accounts, err := contractReader.GetAccountByAK(in.GetAddress())
	if err != nil || accounts == nil {
		reqCtx.GetLog().Warn("QueryAccountContainAK error", "logid", out.Header.Logid, "error", err)
		return out, err
	}

	// get contracts for each account
	out.Contracts = make(map[string]*pb.ContractList)
	for _, account := range accounts {
		contracts, err := contractReader.GetAccountContracts(account)
		if err != nil {
			reqCtx.GetLog().Warn("GetAddressContracts partial account error", "logid", out.Header.Logid, "error", err)
			continue
		}

		if len(contracts) > 0 {
			out.Contracts[account] = &pb.ContractList{
				ContractStatus: contracts,
			}
		}
	}
	return out, nil
}

// DposCandidates get all candidates of the tdpos consensus
func (s *RpcServ) DposCandidates(context.Context, *pb.DposCandidatesRequest) (*pb.DposCandidatesResponse, error) {
	return nil, nil
}

// DposNominateRecords get all records nominated by an user
func (s *RpcServ) DposNominateRecords(context.Context, *pb.DposNominateRecordsRequest) (*pb.DposNominateRecordsResponse, error) {
	return nil, nil
}

// DposNomineeRecords get nominated record of a candidate
func (s *RpcServ) DposNomineeRecords(context.Context, *pb.DposNomineeRecordsRequest) (*pb.DposNomineeRecordsResponse, error) {
	return nil, nil
}

// DposVoteRecords get all vote records voted by an user
func (s *RpcServ) DposVoteRecords(context.Context, *pb.DposVoteRecordsRequest) (*pb.DposVoteRecordsResponse, error) {
	return nil, nil
}

// DposVotedRecords get all vote records of a candidate
func (s *RpcServ) DposVotedRecords(context.Context, *pb.DposVotedRecordsRequest) (*pb.DposVotedRecordsResponse, error) {
	return nil, nil
}

// DposCheckResults get check results of a specific term
func (s *RpcServ) DposCheckResults(context.Context, *pb.DposCheckResultsRequest) (*pb.DposCheckResultsResponse, error) {
	return nil, nil
}

// DposStatus get dpos status
func (s *RpcServ) DposStatus(context.Context, *pb.DposStatusRequest) (*pb.DposStatusResponse, error) {
	return nil, nil
}
