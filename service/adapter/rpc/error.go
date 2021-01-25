package rpc

import (
    "github.com/xuperchain/xupercore/bcs/ledger/xledger/ledger"
    "github.com/xuperchain/xupercore/bcs/ledger/xledger/state"
    "github.com/xuperchain/xupercore/bcs/ledger/xledger/state/utxo"
    common "github.com/xuperchain/xupercore/kernel/engines/xuperos/commom"
    "github.com/xuperchain/xuperos/common/pb"
)

var errorType = map[error]pb.XChainErrorEnum {
    utxo.ErrNoEnoughUTXO:           pb.XChainErrorEnum_NOT_ENOUGH_UTXO_ERROR,
    state.ErrAlreadyInUnconfirmed:  pb.XChainErrorEnum_UTXOVM_ALREADY_UNCONFIRM_ERROR,
    utxo.ErrUTXONotFound:           pb.XChainErrorEnum_UTXOVM_NOT_FOUND_ERROR,
    utxo.ErrInputOutputNotEqual:    pb.XChainErrorEnum_INPUT_OUTPUT_NOT_EQUAL_ERROR,
    utxo.ErrTxNotFound:             pb.XChainErrorEnum_TX_NOT_FOUND_ERROR,
    ledger.ErrTxNotFound:           pb.XChainErrorEnum_TX_NOT_FOUND_ERROR,
    utxo.ErrInvalidSignature:       pb.XChainErrorEnum_TX_SIGN_ERROR,
    common.ErrChainNotExist:        pb.XChainErrorEnum_BLOCKCHAIN_NOTEXIST,
    //pb.XChainErrorEnum_VALIDATE_ERROR,
    //pb.XChainErrorEnum_CANNOT_SYNC_BLOCK_ERROR,
    //pb.XChainErrorEnum_CONFIRM_BLOCK_ERROR,
    //pb.XChainErrorEnum_UTXOVM_PLAY_ERROR,
    //pb.XChainErrorEnum_WALK_ERROR,
    //pb.XChainErrorEnum_NOT_READY_ERROR,
    ledger.ErrBlockNotExist:        pb.XChainErrorEnum_BLOCK_EXIST_ERROR,
    ledger.ErrRootBlockAlreadyExist:pb.XChainErrorEnum_ROOT_BLOCK_EXIST_ERROR,
    ledger.ErrTxDuplicated:         pb.XChainErrorEnum_TX_DUPLICATE_ERROR,
    //pb.XChainErrorEnum_SERVICE_REFUSED_ERROR,
    //pb.XChainErrorEnum_TXDATA_SIGN_ERROR,
    utxo.ErrTxSizeLimitExceeded:    pb.XChainErrorEnum_TX_SLE_ERROR,
    //pb.XChainErrorEnum_TX_FEE_NOT_ENOUGH_ERROR,
    utxo.ErrInvalidSignature:       pb.XChainErrorEnum_UTXO_SIGN_ERROR,
    //pb.XChainErrorEnum_DPOS_QUERY_ERROR,
    utxo.ErrRWSetInvalid:           pb.XChainErrorEnum_RWSET_INVALID_ERROR,
    utxo.ErrInvalidTxExt:           pb.XChainErrorEnum_RWSET_INVALID_ERROR,
    utxo.ErrACLNotEnough:           pb.XChainErrorEnum_RWACL_INVALID_ERROR,
    utxo.ErrGasNotEnough:           pb.XChainErrorEnum_GAS_NOT_ENOUGH_ERROR,
    utxo.ErrVersionInvalid:         pb.XChainErrorEnum_TX_VERSION_INVALID_ERROR,
    //pb.XChainErrorEnum_COMPLIANCE_CHECK_NOT_APPROVED,
    //pb.XChainErrorEnum_ACCOUNT_CONTRACT_STATUS_ERROR,
    //pb.XChainErrorEnum_TX_VERIFICATION_ERROR,
}

// internal error to rpc error
func ErrorEnum(err error) pb.XChainErrorEnum {
    if err == nil {
        return pb.XChainErrorEnum_SUCCESS
    }

    if errorType, ok := errorType[err]; ok {
        return errorType
    }

    return pb.XChainErrorEnum_UNKNOW_ERROR
}
