package rpc

import (
    "strconv"

    "github.com/xuperchain/crypto/core/hash"
    "github.com/xuperchain/xuperchain/core/permission/acl"
    cryptoBase "github.com/xuperchain/xupercore/lib/crypto/client/base"

    "github.com/xuperchain/xuperos/common/pb"
)

func validUtxoAccess(in *pb.UtxoRequest, crypto cryptoBase.CryptoClient, requestAmount int64) bool {
    account := in.GetAddress()
    needLock := in.GetNeedLock()
    if acl.IsAccount(account) == 1 || !needLock {
        return true
    }
    publicKey, err := crypto.GetEcdsaPublicKeyFromJsonStr(in.Publickey)
    if err != nil {
        return false
    }
    checkSignResult, err := crypto.VerifyECDSA(publicKey, in.UserSign, hash.DoubleSha256([]byte(in.Bcname+in.Address+strconv.FormatInt(requestAmount, 10)+strconv.FormatBool(in.NeedLock))))
    if err != nil {
        return false
    }
    if checkSignResult != true {
        return false
    }
    addrMatchCheckResult, _ := crypto.VerifyAddressUsingPublicKey(in.Address, publicKey)
    if addrMatchCheckResult != true {
        return false
    }

    return true
}
