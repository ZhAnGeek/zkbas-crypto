package legendTxTypes

import "math/big"

type DepositTxInfo struct {
	TxType uint8

	// Get from layer1 events.
	AccountNameHash []byte
	AssetId         int64
	AssetAmount     *big.Int

	// Set by layer2.
	AccountIndex int64
}

func (txInfo *DepositTxInfo) GetTxType() int {
	return TxTypeDeposit
}

func (txInfo *DepositTxInfo) Validate() error {
	return nil
}

func (txInfo *DepositTxInfo) VerifySignature(pubKey string) error {
	return nil
}

func (txInfo *DepositTxInfo) GetFromAccountIndex() int64 {
	return NilTxAccountIndex
}

func (txInfo *DepositTxInfo) GetNonce() int64 {
	return NilNonce
}

func (txInfo *DepositTxInfo) GetExpiredAt() int64 {
	return NilExpiredAt
}