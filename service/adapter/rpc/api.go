package rpc

import (
	"context"
	"errors"
	"math/big"
	"strconv"

	ledger "github.com/xuperchain/xupercore/bcs/ledger/xledger/xldgpb"
	"github.com/xuperchain/xupercore/kernel/engines/xuperos/reader"
	"github.com/xuperchain/xupercore/kernel/network/p2p"
	"github.com/xuperchain/xupercore/protos"

	rctx "github.com/xuperchain/xuperos/common/context"
	"github.com/xuperchain/xuperos/common/pb"
)

// PostTx post transaction to blockchain network
func (s *Server) PostTx(ctx context.Context, in *pb.TxStatus) (*pb.CommonReply, error) {
	reqCtx := rctx.ReqCtxFromContext(ctx)
	out := &pb.CommonReply{Header: defRespHeader(in.Header)}

	chain, err := s.engine.Get(in.GetBcname())
	if err != nil {
		out.Header.Error = pb.XChainErrorEnum_BLOCKCHAIN_NOTEXIST
		reqCtx.GetLog().Warn("block chain not exists", "bc", in.GetBcname())
		return out, err
	}

	err = chain.SubmitTx(reqCtx, in.Tx)
	out.Header.Error = ErrorEnum(err)
	if out.Header.Error == pb.XChainErrorEnum_SUCCESS {
		opts := []p2p.MessageOption {
			p2p.WithLogId(in.GetHeader().GetLogid()),
			p2p.WithBCName(in.GetBcname()),
		}
		msg := p2p.NewMessage(protos.XuperMessage_POSTTX, in, opts...)

		engCtx := s.engine.Context()
		go engCtx.Net.SendMessage(reqCtx, msg)
	}

	return out, err
}

// PreExec smart contract preExec process
func (s *Server) PreExec(ctx context.Context, in *pb.InvokeRPCRequest) (*pb.InvokeRPCResponse, error) {
	reqCtx := rctx.ReqCtxFromContext(ctx)
	out := &pb.InvokeRPCResponse{Header: defRespHeader(in.Header)}

	chain, err := s.engine.Get(in.GetBcname())
	if err != nil {
		out.Header.Error = pb.XChainErrorEnum_BLOCKCHAIN_NOTEXIST
		reqCtx.GetLog().Warn("block chain not exists", "bc", in.GetBcname())
		return out, err
	}

	response, err := chain.PreExec(reqCtx, in.GetRequests(), in.GetInitiator(), in.GetAuthRequire())
	if err != nil {
		reqCtx.GetLog().Warn("PreExec error", "error", err)
		return nil, err
	}

	out.Response = response
	return out, nil
}

// PreExecWithSelectUTXO preExec + selectUtxo
func (s *Server) PreExecWithSelectUTXO(ctx context.Context, in *pb.PreExecWithSelectUTXORequest) (*pb.PreExecWithSelectUTXOResponse, error) {
	reqCtx := rctx.ReqCtxFromContext(ctx)
	out := &pb.PreExecWithSelectUTXOResponse{Header: defRespHeader(in.Header), Bcname: in.GetBcname()}

	chain, err := s.engine.Get(in.GetBcname())
	if err != nil {
		out.Header.Error = pb.XChainErrorEnum_BLOCKCHAIN_NOTEXIST
		reqCtx.GetLog().Warn("block chain not exists", "bc", in.GetBcname())
		return out, err
	}

	// PreExec
	preExecRequest := in.GetRequest()
	fee := int64(0)
	if preExecRequest != nil {
		preExecRequest.Header = in.Header
		invokeRPCResponse, err := s.PreExec(ctx, preExecRequest)
		if err != nil {
			return nil, err
		}
		invokeResponse := invokeRPCResponse.GetResponse()
		out.Response = invokeResponse
		fee = out.Response.GetGasUsed()
	}

	// SelectUTXO
	totalAmount := in.GetTotalAmount() + fee
	if totalAmount > 0 {
		utxoInput := &pb.UtxoRequest {
			Bcname:    in.GetBcname(),
			Address:   in.GetAddress(),
			TotalNeed: strconv.FormatInt(totalAmount, 10),
			Publickey: in.GetSignInfo().GetPublicKey(),
			UserSign:  in.GetSignInfo().GetSign(),
			NeedLock:  in.GetNeedLock(),
		}

		if ok := validUtxoAccess(utxoInput, chain.Context().Crypto, in.GetTotalAmount()); !ok {
			return nil, errors.New("validUtxoAccess failed")
		}

		utxoOutput, err := s.SelectUTXO(ctx, utxoInput)
		if err != nil {
			return nil, err
		}

		out.UtxoOutput = &ledger.UtxoOutput{
			UtxoList: utxoOutput.UtxoList,
			TotalSelected: utxoOutput.TotalSelected,
		}
	}

	return out, nil
}

// SelectUTXO select utxo inputs depending on amount
func (s *Server) SelectUTXO(ctx context.Context, in *pb.UtxoRequest) (*pb.UtxoResponse, error) {
	reqCtx := rctx.ReqCtxFromContext(ctx)
	out := &pb.UtxoResponse{Header: defRespHeader(in.Header)}

	totalNeed, ok := new(big.Int).SetString(in.TotalNeed, 10)
	if !ok {
		out.Header.Error = pb.XChainErrorEnum_CONNECT_REFUSE // 拒绝
		return out, nil
	}

	chain, err := s.engine.Get(in.GetBcname())
	if err != nil {
		out.Header.Error = pb.XChainErrorEnum_BLOCKCHAIN_NOTEXIST
		reqCtx.GetLog().Warn("failed to select utxo, bcname not exists", "bcName", in.GetBcname())
		return out, err
	}

	utxoReader := reader.NewUtxoReader(chain.Context(), reqCtx)
	response, err := utxoReader.SelectUTXO(in.GetAddress(), totalNeed, in.GetNeedLock(), false)
	if err != nil {
		out.Header.Error = ErrorEnum(err)
		reqCtx.GetLog().Warn("failed to select utxo", "error", err)
		return out, err
	}

	out.UtxoList = response.UtxoList
	out.TotalSelected = response.TotalSelected
	reqCtx.GetLog().SetInfoField("totalSelect", out.TotalSelected)
	return out, nil
}

// SelectUTXOBySize select utxo inputs depending on size
func (s *Server) SelectUTXOBySize(ctx context.Context, in *pb.UtxoRequest) (*pb.UtxoResponse, error) {
	reqCtx := rctx.ReqCtxFromContext(ctx)
	out := &pb.UtxoResponse{Header: defRespHeader(in.Header)}

	chain, err := s.engine.Get(in.GetBcname())
	if err != nil {
		out.Header.Error = pb.XChainErrorEnum_BLOCKCHAIN_NOTEXIST
		reqCtx.GetLog().Warn("failed to merge utxo, bcname not exists", "logid", in.Header.Logid)
		return out, err
	}

	utxoReader := reader.NewUtxoReader(chain.Context(), reqCtx)
	response, err := utxoReader.SelectUTXOBySize(in.GetAddress(), in.GetNeedLock(), false)
	if err != nil {
		out.Header.Error = ErrorEnum(err)
		reqCtx.GetLog().Warn("failed to select utxo", "error", err)
		return out, err
	}

	out.UtxoList = response.UtxoList
	out.TotalSelected = response.TotalSelected
	reqCtx.GetLog().SetInfoField("totalSelect", out.TotalSelected)
	return out, nil
}

// QueryContractStatData query statistic info about contract
func (s *Server) QueryContractStatData(ctx context.Context, in *pb.ContractStatDataRequest) (*pb.ContractStatDataResponse, error) {
	reqCtx := rctx.ReqCtxFromContext(ctx)
	out := &pb.ContractStatDataResponse{Header: defRespHeader(in.Header)}

	chain, err := s.engine.Get(in.GetBcname())
	if err != nil {
		out.Header.Error = pb.XChainErrorEnum_BLOCKCHAIN_NOTEXIST
		reqCtx.GetLog().Warn("block chain not exists", "bc", in.GetBcname())
		return out, err
	}

	contractReader := reader.NewContractReader(chain.Context(), reqCtx)
	contractStatData, err := contractReader.QueryContractStatData()
	if err != nil {
		return nil, err
	}

	out.Data = contractStatData
	return out, nil
}

// QueryUtxoRecord query utxo records
func (s *Server) QueryUtxoRecord(ctx context.Context, in *pb.UtxoRecordDetails) (*pb.UtxoRecordDetails, error) {
	reqCtx := rctx.ReqCtxFromContext(ctx)
	out := &pb.UtxoRecordDetails{Header: defRespHeader(in.Header)}

	chain, err := s.engine.Get(in.GetBcname())
	if err != nil {
		out.Header.Error = pb.XChainErrorEnum_BLOCKCHAIN_NOTEXIST
		reqCtx.GetLog().Warn("block chain not exists", "bc", in.GetBcname())
		return out, err
	}

	utxoReader := reader.NewUtxoReader(chain.Context(), reqCtx)
	if len(in.GetAccountName()) > 0 {
		utxoRecord, err := utxoReader.QueryUtxoRecord(in.GetAccountName(), in.GetDisplayCount())
		if err != nil {
			reqCtx.GetLog().Warn("query utxo record error", "account", in.GetAccountName())
			return out, err
		}

		out.FrozenUtxoRecord = utxoRecord.FrozenUtxo
		out.LockedUtxoRecord =  utxoRecord.LockedUtxo
		out.OpenUtxoRecord = utxoRecord.OpenUtxo
		return out, nil
	}

	return out, nil
}

// QueryACL query some account info
func (s *Server) QueryACL(ctx context.Context, in *pb.AclStatus) (*pb.AclStatus, error) {
	reqCtx := rctx.ReqCtxFromContext(ctx)
	out := &pb.AclStatus{Header: defRespHeader(in.Header)}

	chain, err := s.engine.Get(in.GetBcname())
	if err != nil {
		out.Header.Error = pb.XChainErrorEnum_BLOCKCHAIN_NOTEXIST
		reqCtx.GetLog().Warn("block chain not exists", "bc", in.GetBcname())
		return out, err
	}

	contractReader := reader.NewContractReader(chain.Context(), reqCtx)
	accountName := in.GetAccountName()
	contractName := in.GetContractName()
	methodName := in.GetMethodName()
	if len(accountName) > 0 {
		acl, err := contractReader.QueryAccountACL(accountName)
		if err != nil {
			out.Confirmed = false
			reqCtx.GetLog().Warn("query account acl error", "account", accountName)
			return out, err
		}
		out.Confirmed = true
		out.Acl = acl
	} else if len(contractName) > 0 {
		if len(methodName) > 0 {
			acl, err := contractReader.QueryContractMethodACL(contractName, methodName)
			if err != nil {
				out.Confirmed = false
				reqCtx.GetLog().Warn("query contract method acl error", "account", accountName, "method", methodName)
				return out, err
			}
			out.Confirmed = true
			out.Acl = acl
		}
	}
	return out, nil
}

// GetAccountContracts get account request
func (s *Server) GetAccountContracts(ctx context.Context, in *pb.GetAccountContractsRequest) (*pb.GetAccountContractsResponse, error) {
	reqCtx := rctx.ReqCtxFromContext(ctx)
	out := &pb.GetAccountContractsResponse{Header: defRespHeader(in.Header)}

	chain, err := s.engine.Get(in.GetBcname())
	if err != nil {
		out.Header.Error = pb.XChainErrorEnum_BLOCKCHAIN_NOTEXIST
		reqCtx.GetLog().Warn("block chain not exists", "bc", in.GetBcname())
		return out, err
	}

	contractReader := reader.NewContractReader(chain.Context(), reqCtx)
	contractsStatus, err := contractReader.GetAccountContracts(in.GetAccount())
	if err != nil {
		out.Header.Error = pb.XChainErrorEnum_ACCOUNT_CONTRACT_STATUS_ERROR
		reqCtx.GetLog().Warn("GetAccountContracts error", "error", err)
		return out, err
	}
	out.ContractsStatus = contractsStatus
	return out, nil
}

// QueryTx Get transaction details
func (s *Server) QueryTx(ctx context.Context, in *pb.TxStatus) (*pb.TxStatus, error) {
	reqCtx := rctx.ReqCtxFromContext(ctx)
	out := &pb.TxStatus{Header: defRespHeader(in.Header)}

	chain, err := s.engine.Get(in.GetBcname())
	if err != nil {
		out.Header.Error = pb.XChainErrorEnum_BLOCKCHAIN_NOTEXIST
		reqCtx.GetLog().Warn("block chain not exists", "bc", in.GetBcname())
		return out, err
	}

	ledgerReader := reader.NewLedgerReader(chain.Context(), reqCtx)
	txInfo, err := ledgerReader.QueryTx(in.GetTxid())
	if err != nil {
		reqCtx.GetLog().Warn("query tx error", "txid", in.GetTxid())
		return out, err
	}

	out.Tx = txInfo.Tx
	out.Status = txInfo.Status
	out.Distance = txInfo.Distance
	return out, nil
}

// GetBalance get balance for account or addr
func (s *Server) GetBalance(ctx context.Context, in *pb.AddressStatus) (*pb.AddressStatus, error) {
	reqCtx := rctx.ReqCtxFromContext(ctx)

	for i := 0; i < len(in.Bcs); i++ {
		chain, err := s.engine.Get(in.Bcs[i].Bcname)
		if err != nil {
			in.Bcs[i].Error = pb.XChainErrorEnum_BLOCKCHAIN_NOTEXIST
			in.Bcs[i].Balance = ""
			continue
		}

		utxoReader := reader.NewUtxoReader(chain.Context(), reqCtx)
		balance, err := utxoReader.GetBalance(in.Address)
		if err != nil {
			in.Bcs[i].Error = ErrorEnum(err)
			in.Bcs[i].Balance = ""
		} else {
			in.Bcs[i].Error = pb.XChainErrorEnum_SUCCESS
			in.Bcs[i].Balance = balance
		}
	}
	return in, nil
}

// GetFrozenBalance get balance frozened for account or addr
func (s *Server) GetFrozenBalance(ctx context.Context, in *pb.AddressStatus) (*pb.AddressStatus, error) {
	reqCtx := rctx.ReqCtxFromContext(ctx)

	for i := 0; i < len(in.Bcs); i++ {
		chain, err := s.engine.Get(in.Bcs[i].Bcname)
		if err != nil {
			in.Bcs[i].Error = pb.XChainErrorEnum_BLOCKCHAIN_NOTEXIST
			in.Bcs[i].Balance = ""
			continue
		}

		utxoReader := reader.NewUtxoReader(chain.Context(), reqCtx)
		balance, err := utxoReader.GetFrozenBalance(in.Address)
		if err != nil {
			in.Bcs[i].Error = ErrorEnum(err)
			in.Bcs[i].Balance = ""
		} else {
			in.Bcs[i].Error = pb.XChainErrorEnum_SUCCESS
			in.Bcs[i].Balance = balance
		}
	}

	return in, nil
}

// GetBalanceDetail get balance frozened for account or addr
func (s *Server) GetBalanceDetail(ctx context.Context, in *pb.AddressBalanceStatus) (*pb.AddressBalanceStatus, error) {
	reqCtx := rctx.ReqCtxFromContext(ctx)

	for i := 0; i < len(in.Tfds); i++ {
		chain, err := s.engine.Get(in.Tfds[i].Bcname)
		if err != nil {
			in.Tfds[i].Error = pb.XChainErrorEnum_BLOCKCHAIN_NOTEXIST
			in.Tfds[i].Tfd = nil
			continue
		}

		utxoReader := reader.NewUtxoReader(chain.Context(), reqCtx)
		tfd, err := utxoReader.GetBalanceDetail(in.Address)
		if err != nil {
			in.Tfds[i].Error = ErrorEnum(err)
			in.Tfds[i].Tfd = nil
		} else {
			in.Tfds[i].Error = pb.XChainErrorEnum_SUCCESS
			// TODO: 使用了ledger定义的类型，验证是否有效
			in.Tfds[i].Tfd = tfd
		}
	}

	return in, nil
}

// GetBlock get block info according to blockID
func (s *Server) GetBlock(ctx context.Context, in *pb.BlockID) (*pb.Block, error) {
	reqCtx := rctx.ReqCtxFromContext(ctx)
	out := &pb.Block{Header: defRespHeader(in.Header)}

	chain, err := s.engine.Get(in.GetBcname())
	if err != nil {
		out.Header.Error = pb.XChainErrorEnum_BLOCKCHAIN_NOTEXIST
		reqCtx.GetLog().Warn("block chain not exists", "bc", in.GetBcname())
		return out, err
	}

	ledgerReader := reader.NewLedgerReader(chain.Context(), reqCtx)
	blockInfo, err := ledgerReader.QueryBlock(in.Blockid, true)
	if err != nil {
		reqCtx.GetLog().Warn("query block error", "error", err)
		return out, nil
	}

	// 类型转换：ledger.BlockInfo => pb.Block
	out.Block = blockInfo.Block
	out.Status = pb.Block_EBlockStatus(blockInfo.Status)

	block := blockInfo.GetBlock()
	transactions := block.GetTransactions()
	transactionsFilter := make([]*ledger.Transaction, 0, len(transactions))
	for _, transaction := range transactions {
		transactionsFilter = append(transactionsFilter, transaction)
	}

	if transactions != nil {
		out.Block.Transactions = transactionsFilter
	}

	reqCtx.GetLog().SetInfoField("blockid", out.GetBlockid())
	reqCtx.GetLog().SetInfoField("height", out.GetBlock().GetHeight())
	return out, nil
}

// GetBlockChainStatus get systemstatus
func (s *Server) GetBlockChainStatus(ctx context.Context, in *pb.BCStatus) (*pb.BCStatus, error) {
	reqCtx := rctx.ReqCtxFromContext(ctx)
	out := &pb.BCStatus{Header: defRespHeader(in.Header)}

	chain, err := s.engine.Get(in.GetBcname())
	if err != nil {
		out.Header.Error = pb.XChainErrorEnum_BLOCKCHAIN_NOTEXIST
		reqCtx.GetLog().Warn("block chain not exists", "bc", in.GetBcname())
		return out, err
	}

	chainReader := reader.NewChainReader(chain.Context(), reqCtx)
	status, err := chainReader.GetChainStatus()
	if err != nil {
		reqCtx.GetLog().Warn("get chain status error", "error", err)
	}

	// 类型转换：=> pb.BCStatus
	out.Meta = status.LedgerMeta
	out.Block = status.Block
	out.UtxoMeta = status.UtxoMeta
	return out, nil
}

// ConfirmBlockChainStatus confirm is_trunk
func (s *Server) ConfirmBlockChainStatus(ctx context.Context, in *pb.BCStatus) (*pb.BCTipStatus, error) {
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
func (s *Server) GetBlockChains(ctx context.Context, in *pb.CommonIn) (*pb.BlockChains, error) {
	out := &pb.BlockChains{Header: defRespHeader(in.Header)}
	out.Blockchains = s.engine.GetChains()
	return out, nil
}

// GetSystemStatus get systemstatus
func (s *Server) GetSystemStatus(ctx context.Context, in *pb.CommonIn) (*pb.SystemsStatusReply, error) {
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
func (s *Server) GetNetURL(ctx context.Context, in *pb.CommonIn) (*pb.RawUrl, error) {
	out := &pb.RawUrl{Header: defRespHeader(in.Header)}
	peerInfo := s.engine.Context().Net.PeerInfo()
	out.RawUrl = peerInfo.Address
	return out, nil
}

// GetBlockByHeight  get trunk block by height
func (s *Server) GetBlockByHeight(ctx context.Context, in *pb.BlockHeight) (*pb.Block, error) {
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
func (s *Server) GetAccountByAK(ctx context.Context, in *pb.AK2AccountRequest) (*pb.AK2AccountResponse, error) {
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
func (s *Server) GetAddressContracts(ctx context.Context, in *pb.AddressContractsRequest) (*pb.AddressContractsResponse, error) {
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
//
////  DposCandidates get all candidates of the tdpos consensus
//func (s *Server) DposCandidates(context.Context, *pb.DposCandidatesRequest) (*pb.DposCandidatesResponse, error) {
//	return nil, nil
//}
////  DposNominateRecords get all records nominated by an user
//func (s *Server) DposNominateRecords(context.Context, *pb.DposNominateRecordsRequest) (*pb.DposNominateRecordsResponse, error){
//	return nil, nil
//}
////  DposNomineeRecords get nominated record of a candidate
//func (s *Server) DposNomineeRecords(context.Context, *pb.DposNomineeRecordsRequest) (*pb.DposNomineeRecordsResponse, error){
//	return nil, nil
//}
////  DposVoteRecords get all vote records voted by an user
//func (s *Server) DposVoteRecords(context.Context, *pb.DposVoteRecordsRequest) (*pb.DposVoteRecordsResponse, error){
//	return nil, nil
//}
////  DposVotedRecords get all vote records of a candidate
//func (s *Server) DposVotedRecords(context.Context, *pb.DposVotedRecordsRequest) (*pb.DposVotedRecordsResponse, error){
//	return nil, nil
//}
////  DposCheckResults get check results of a specific term
//func (s *Server) DposCheckResults(context.Context, *pb.DposCheckResultsRequest) (*pb.DposCheckResultsResponse, error){
//	return nil, nil
//}
//// DposStatus get dpos status
//func (s *Server) DposStatus(context.Context, *pb.DposStatusRequest) (*pb.DposStatusResponse, error){
//	return nil, nil
//}
