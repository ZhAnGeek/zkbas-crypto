package test

import (
	"encoding/json"
	"github.com/bnb-chain/zkbnb-crypto/wasm/legend/legendTxTypes"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestWithdrawSegmentFormat(t *testing.T) {

	var segmentFormat *legendTxTypes.WithdrawSegmentFormat
	segmentFormat = &legendTxTypes.WithdrawSegmentFormat{
		FromAccountIndex:  0,
		AssetId:           0,
		AssetAmount:       "100",
		GasAccountIndex:   1,
		GasFeeAssetId:     3,
		GasFeeAssetAmount: "3",
		ToAddress:         "0x507Bd54B4232561BC0Ca106F7b029d064fD6bE4c",
		ExpiredAt:         1654656781000, // milli seconds
		Nonce:             1,
	}

	res, err := json.Marshal(segmentFormat)
	assert.Nil(t, err)

	log.Println(string(res))
}
