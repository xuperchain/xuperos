package models

import (
	"math/big"

	lpb "github.com/xuperchain/xupercore/bcs/ledger/xledger/xldgpb"
	xctx "github.com/xuperchain/xupercore/kernel/common/xcontext"
	ecom "github.com/xuperchain/xupercore/kernel/engines/xuperos/common"
	"github.com/xuperchain/xupercore/kernel/engines/xuperos/reader"
	"github.com/xuperchain/xupercore/kernel/engines/xuperos/xpb"
	cryptoBase "github.com/xuperchain/xupercore/lib/crypto/client/base"
	"github.com/xuperchain/xupercore/lib/logs"
	"github.com/xuperchain/xupercore/protos"

	sctx "github.com/xuperchain/xuperos/common/context"
)

type ChainHandle struct {
	bcName string
	reqCtx sctx.ReqCtx
	log    logs.Logger
	chain  ecom.Chain
}

func NewChainHandle(bcName string, reqCtx sctx.ReqCtx) (*ChainHandle, error) {
	if bcName == "" || reqCtx == nil || reqCtx.GetEngine() == nil {
		return nil, ecom.ErrParameter
	}

	chain, err := reqCtx.GetEngine().Get(bcName)
	if err != nil {
		return nil, ecom.ErrChainNotExist
	}

	obj := &ChainHandle{
		bcName: bcName,
		reqCtx: reqCtx,
		log:    reqCtx.GetLog(),
		chain:  chain,
	}
	return obj, nil
}

func (t *ChainHandle) GetCrypto() *cryptoBase.CryptoClient {
	return t.chain.Context().Crypto
}

func (t *ChainHandle) SubmitTx(tx *lpb.Transaction) error {
	return t.chain.SubmitTx(t.genXctx(), tx)
}

func (t *ChainHandle) PreExec(req []*protos.InvokeRequest,
	initiator string, authRequires []string) (*protos.InvokeResponse, error) {
	return t.chain.PreExec(t.genXctx(), req, initiator, authRequires)
}

func (t *ChainHandle) QueryTx(txId []byte) (*xpb.TxInfo, error) {
	return reader.NewLedgerReader(t.chain.Context(), t.genXctx()).QueryTx(txId)
}

func (t *ChainHandle) SelectUtxo(account string, need *big.Int, isLock, isExclude bool,
	pubKey string, sign []byte) (*lpb.UtxoOutput, error) {
	// 如果需要临时锁定utxo，需要校验权限
	ok := t.checkSelectUtxoSign(account, pubKey, sign, isLock, need)
	if !ok {
		rctx.GetLog().Warn("select utxo verify sign failed", "account", account, "isLock", isLock)
		return resp, ecom.ErrUnauthorized
	}

	return reader.NewUtxoReader(t.chain.Context(), t.genXctx()).SelectUTXO(account, need,
		isLock, isExclude)
}

func (t *ChainHandle) QueryBlock(blkId []byte, needContent bool) (*xpb.BlockInfo, error) {
	return reader.NewLedgerReader(t.chain.Context(), t.genXctx()).QueryBlock(blkId, needContent)
}

func (t *ChainHandle) QueryChainStatus(needBranch bool) (*xpb.ChainStatus, error) {
	return reader.NewChainReader(t.chain.Context(), t.genXctx()).GetChainStatus()
}

func (t *ChainHandle) genXctx() xctx.XContext {
	return &xctx.BaseCtx{
		XLog:  t.reqCtx.GetLog(),
		Timer: t.reqCtx.GetTimer(),
	}
}

func (t *ChainHandle) checkSelectUtxoSign(account, pubKey string, sign []byte,
	isLock bool, need *big.Int) bool {
	// 只对需要临时锁定utxo的校验
	if aclUtils.IsAccount(account) == 1 || !needLock {
		return true
	}

	crypto := t.chain.Context().Crypto
	publicKey, err := crypto.GetEcdsaPublicKeyFromJsonStr(pubKey)
	if err != nil {
		return false
	}

	hashStr := t.bcName + account + need.String() + strconv.FormatBool(isLock)
	doubleHash := cryptoHash.DoubleSha256([]byte(hashStr))
	checkSignResult, err := crypto.VerifyECDSA(publicKey, sign, doubleHash)
	if err != nil {
		return false
	}
	if checkSignResult != true {
		return false
	}
	addrMatchCheckResult, _ := crypto.VerifyAddressUsingPublicKey(account, publicKey)
	if addrMatchCheckResult != true {
		return false
	}

	return true
}
