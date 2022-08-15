/*
 * Copyright © 2021 Zecrey Protocol
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package zero

import (
	"bytes"
	"errors"
	curve "github.com/bnb-chain/zkbas-crypto/ecc/ztwistededwards/tebn254"
	"github.com/bnb-chain/zkbas-crypto/elgamal/twistededwards/tebn254/twistedElgamal"
	"github.com/bnb-chain/zkbas-crypto/ffmath"
	"github.com/bnb-chain/zkbas-crypto/hash/bn254/zmimc"
	"github.com/bnb-chain/zkbas-crypto/util"
	"log"
	"math"
	"math/big"
)

func ProveAddLiquidity(relation *AddLiquidityRelation) (proof *AddLiquidityProof, err error) {
	if relation == nil {
		log.Println("[ProveAddLiquidity] invalid params")
		return nil, errors.New("[ProveAddLiquidity] invalid params")
	}
	var (
		alpha_r_DeltaLP                                         *big.Int
		A_CLPL_Delta, A_CLPR_DeltaHExp_DeltaLPNeg               *Point
		alpha_sk_u, alpha_sk_uInv, alpha_bar_r_A, alpha_bar_r_B *big.Int
		A_pk_u, A_T_uAC_uARPrimeInv, A_T_uBC_uBRPrimeInv        *Point
		c                                                       *big.Int
		Z_rDelta_LP                                             *big.Int
		Z_sk_u, Z_bar_r_A, Z_bar_r_B, Z_sk_uInv                 *big.Int
		buf                                                     bytes.Buffer
		// gas part
		A_T_feeC_feeRPrimeInv *Point
		Z_bar_r_fee           *big.Int
		C_feeLPrimeInv        *Point
		alpha_bar_r_fee       *big.Int
	)
	buf.Write(PaddingBigIntBytes(FixedCurve))
	writePointIntoBuf(&buf, relation.Pk_u)
	writePointIntoBuf(&buf, relation.Pk_pool)
	writeEncIntoBuf(&buf, relation.C_uA)
	writeEncIntoBuf(&buf, relation.C_uB)
	writeEncIntoBuf(&buf, relation.C_uA_Delta)
	writeEncIntoBuf(&buf, relation.C_uB_Delta)
	writeEncIntoBuf(&buf, relation.C_LP_Delta)
	writePointIntoBuf(&buf, relation.T_uA)
	writePointIntoBuf(&buf, relation.T_uB)
	// assets id
	writeUint64IntoBuf(&buf, uint64(relation.AssetAId))
	writeUint64IntoBuf(&buf, uint64(relation.AssetBId))
	// valid enc
	alpha_r_DeltaLP = curve.RandomValue()
	A_CLPL_Delta = curve.ScalarMul(relation.Pk_u, alpha_r_DeltaLP)
	A_CLPR_DeltaHExp_DeltaLPNeg = curve.ScalarMul(G, alpha_r_DeltaLP)
	// write into buf
	writePointIntoBuf(&buf, A_CLPL_Delta)
	writePointIntoBuf(&buf, A_CLPR_DeltaHExp_DeltaLPNeg)
	// ownership
	alpha_sk_u = curve.RandomValue()
	alpha_sk_uInv = ffmath.ModInverse(alpha_sk_u, Order)
	alpha_bar_r_A = curve.RandomValue()
	alpha_bar_r_B = curve.RandomValue()
	A_pk_u = curve.ScalarMul(G, alpha_sk_u)
	// user asset A part
	A_T_uAC_uARPrimeInv = curve.Add(
		relation.C_uA.CL,
		relation.C_uA_Delta.CL,
	)
	A_T_uAC_uARPrimeInv = curve.Neg(A_T_uAC_uARPrimeInv)
	A_T_uAC_uARPrimeInv = curve.ScalarMul(A_T_uAC_uARPrimeInv, alpha_sk_uInv)
	A_T_uAC_uARPrimeInv = curve.Add(
		A_T_uAC_uARPrimeInv,
		curve.ScalarMul(G, alpha_bar_r_A))
	// user asset B part
	A_T_uBC_uBRPrimeInv = curve.Add(
		relation.C_uB.CL,
		relation.C_uB_Delta.CL,
	)
	A_T_uBC_uBRPrimeInv = curve.Neg(A_T_uBC_uBRPrimeInv)
	A_T_uBC_uBRPrimeInv = curve.ScalarMul(A_T_uBC_uBRPrimeInv, alpha_sk_uInv)
	A_T_uBC_uBRPrimeInv = curve.Add(A_T_uBC_uBRPrimeInv, curve.ScalarMul(G, alpha_bar_r_B))
	if relation.GasFeeAssetId == relation.AssetAId {
		// gas part
		alpha_bar_r_fee = new(big.Int).Set(alpha_bar_r_A)
		A_T_feeC_feeRPrimeInv = new(Point).Set(A_T_uAC_uARPrimeInv)
	} else if relation.GasFeeAssetId == relation.AssetBId {
		// gas part
		alpha_bar_r_fee = new(big.Int).Set(alpha_bar_r_B)
		A_T_feeC_feeRPrimeInv = new(Point).Set(A_T_uBC_uBRPrimeInv)
	} else {
		// gas part
		C_feeLPrimeInv = curve.Neg(relation.C_fee.CL)
		alpha_bar_r_fee = curve.RandomValue()
		A_T_feeC_feeRPrimeInv = curve.Add(curve.ScalarMul(G, alpha_bar_r_fee), curve.ScalarMul(C_feeLPrimeInv, alpha_sk_uInv))
	}
	// gas fee
	writePointIntoBuf(&buf, A_T_feeC_feeRPrimeInv)
	writeEncIntoBuf(&buf, relation.C_fee)
	writeUint64IntoBuf(&buf, uint64(relation.GasFeeAssetId))
	writeUint64IntoBuf(&buf, relation.GasFee)
	// write into buf
	writePointIntoBuf(&buf, A_pk_u)
	writePointIntoBuf(&buf, A_T_uAC_uARPrimeInv)
	writePointIntoBuf(&buf, A_T_uBC_uBRPrimeInv)
	// compute challenge
	c, err = util.HashToInt(buf, zmimc.Hmimc)
	if err != nil {
		return nil, err
	}
	// compute response values
	Z_rDelta_LP = ffmath.AddMod(alpha_r_DeltaLP, ffmath.Multiply(c, relation.R_DeltaLP), Order)
	Z_sk_u = ffmath.AddMod(alpha_sk_u, ffmath.Multiply(c, relation.Sk_u), Order)
	Z_bar_r_A = ffmath.AddMod(alpha_bar_r_A, ffmath.Multiply(c, relation.Bar_r_A), Order)
	Z_bar_r_B = ffmath.AddMod(alpha_bar_r_B, ffmath.Multiply(c, relation.Bar_r_B), Order)
	Z_sk_uInv = ffmath.AddMod(alpha_sk_uInv, ffmath.Multiply(c, ffmath.ModInverse(relation.Sk_u, Order)), Order)
	// gas fee
	Z_bar_r_fee = ffmath.AddMod(alpha_bar_r_fee, ffmath.Multiply(c, relation.Bar_r_fee), Order)
	// constrcut proof
	proof = &AddLiquidityProof{
		A_CLPL_Delta:                A_CLPL_Delta,
		A_CLPR_DeltaHExp_DeltaLPNeg: A_CLPR_DeltaHExp_DeltaLPNeg,
		Z_rDelta_LP:                 Z_rDelta_LP,
		A_pk_u:                      A_pk_u,
		A_T_uAC_uARPrimeInv:         A_T_uAC_uARPrimeInv,
		A_T_uBC_uBRPrimeInv:         A_T_uBC_uBRPrimeInv,
		Z_sk_u:                      Z_sk_u,
		Z_bar_r_A:                   Z_bar_r_A,
		Z_bar_r_B:                   Z_bar_r_B,
		Z_sk_uInv:                   Z_sk_uInv,
		ARangeProof:                 relation.ARangeProof,
		BRangeProof:                 relation.BRangeProof,
		C_uA:                        relation.C_uA,
		C_uB:                        relation.C_uB,
		C_uA_Delta:                  relation.C_uA_Delta,
		C_uB_Delta:                  relation.C_uB_Delta,
		LC_poolA_Delta:              relation.LC_poolA_Delta,
		LC_poolB_Delta:              relation.LC_poolB_Delta,
		C_LP_Delta:                  relation.C_LP_Delta,
		Pk_u:                        relation.Pk_u,
		Pk_pool:                     relation.Pk_pool,
		R_DeltaA:                    relation.R_DeltaA,
		R_DeltaB:                    relation.R_DeltaB,
		T_uA:                        relation.T_uA,
		T_uB:                        relation.T_uB,
		B_poolA:                     relation.B_poolA,
		B_poolB:                     relation.B_poolB,
		B_A_Delta:                   relation.B_A_Delta,
		B_B_Delta:                   relation.B_B_Delta,
		Delta_LP:                    relation.Delta_LP,
		AssetAId:                    relation.AssetAId,
		AssetBId:                    relation.AssetBId,
		A_T_feeC_feeRPrimeInv:       A_T_feeC_feeRPrimeInv,
		Z_bar_r_fee:                 Z_bar_r_fee,
		C_fee:                       relation.C_fee,
		T_fee:                       relation.T_fee,
		GasFeeAssetId:               relation.GasFeeAssetId,
		GasFee:                      relation.GasFee,
		GasFeePrimeRangeProof:       relation.GasFeePrimeRangeProof,
	}
	return proof, nil
}

func (proof *AddLiquidityProof) Verify() (res bool, err error) {
	// verify range proofs
	if !proof.ARangeProof.A.Equal(proof.T_uA) || !proof.BRangeProof.A.Equal(proof.T_uB) {
		log.Println("[Verify AddLiquidityProof] invalid params")
		return false, errors.New("[Verify AddLiquidityProof] invalid params")
	}
	var (
		C_uAPrime, C_uBPrime       *ElGamalEnc
		C_uAPrimeNeg, C_uBPrimeNeg *ElGamalEnc
		c                          *big.Int
		buf                        bytes.Buffer
	)
	// challenge buf
	buf.Write(PaddingBigIntBytes(FixedCurve))
	writePointIntoBuf(&buf, proof.Pk_u)
	writePointIntoBuf(&buf, proof.Pk_pool)
	writeEncIntoBuf(&buf, proof.C_uA)
	writeEncIntoBuf(&buf, proof.C_uB)
	writeEncIntoBuf(&buf, proof.C_uA_Delta)
	writeEncIntoBuf(&buf, proof.C_uB_Delta)
	writeEncIntoBuf(&buf, proof.C_LP_Delta)
	writePointIntoBuf(&buf, proof.T_uA)
	writePointIntoBuf(&buf, proof.T_uB)
	// assets id
	writeUint64IntoBuf(&buf, uint64(proof.AssetAId))
	writeUint64IntoBuf(&buf, uint64(proof.AssetBId))
	// write into buf
	writePointIntoBuf(&buf, proof.A_CLPL_Delta)
	writePointIntoBuf(&buf, proof.A_CLPR_DeltaHExp_DeltaLPNeg)
	// write into buf
	// gas fee
	writePointIntoBuf(&buf, proof.A_T_feeC_feeRPrimeInv)
	writeEncIntoBuf(&buf, proof.C_fee)
	writeUint64IntoBuf(&buf, uint64(proof.GasFeeAssetId))
	writeUint64IntoBuf(&buf, proof.GasFee)

	writePointIntoBuf(&buf, proof.A_pk_u)
	writePointIntoBuf(&buf, proof.A_T_uAC_uARPrimeInv)
	writePointIntoBuf(&buf, proof.A_T_uBC_uBRPrimeInv)
	// compute challenge
	c, err = util.HashToInt(buf, zmimc.Hmimc)
	if err != nil {
		log.Println("[Verify AddLiquidityProof] unable to compute challenge")
		return false, err
	}
	// verify params
	isValidParams, err := verifyAddLiquidityParams(proof)
	if err != nil {
		return false, err
	}
	if !isValidParams {
		return false, errors.New("[Verify AddLiquidityProof] invalid params")
	}
	// verify enc
	l1 := curve.ScalarMul(proof.Pk_u, proof.Z_rDelta_LP)
	r1 := curve.Add(proof.A_CLPL_Delta, curve.ScalarMul(proof.C_LP_Delta.CL, c))
	if !l1.Equal(r1) {
		log.Println("[Verify AddLiquidityProof] l1 != r1")
		return false, nil
	}
	// verify ownership
	l2 := curve.ScalarMul(G, proof.Z_sk_u)
	r2 := curve.Add(proof.A_pk_u, curve.ScalarMul(proof.Pk_u, c))
	if !l2.Equal(r2) {
		log.Println("[Verify AddLiquidityProof] l2 != r2")
		return false, nil
	}
	if proof.GasFeeAssetId == proof.AssetAId {
		if !equalEnc(proof.C_uA, proof.C_fee) || !proof.A_T_uAC_uARPrimeInv.Equal(proof.A_T_feeC_feeRPrimeInv) {
			log.Println("[Verify AddLiquidityProof] invalid params")
			return false, errors.New("[Verify AddLiquidityProof] invalid params")
		}
		// A part
		C_uAPrime, err = twistedElgamal.EncAdd(proof.C_uA, proof.C_uA_Delta)
		if err != nil {
			return false, err
		}
		C_uAPrime.CR = curve.Add(C_uAPrime.CR, curve.ScalarMul(H, big.NewInt(-int64(proof.GasFee))))
		C_uAPrimeNeg = negElgamal(C_uAPrime)
		l3 := curve.Add(
			curve.ScalarMul(G, proof.Z_bar_r_A),
			curve.ScalarMul(C_uAPrimeNeg.CL, proof.Z_sk_uInv),
		)
		r3 := curve.Add(
			proof.A_T_uAC_uARPrimeInv,
			curve.ScalarMul(
				curve.Add(
					proof.T_uA,
					C_uAPrimeNeg.CR,
				),
				c,
			),
		)
		if !l3.Equal(r3) {
			log.Println("[Verify AddLiquidityProof] l3 != r3")
			return false, nil
		}
		// B paart
		C_uBPrime, err = twistedElgamal.EncAdd(proof.C_uB, proof.C_uB_Delta)
		if err != nil {
			return false, err
		}
		C_uBPrimeNeg = negElgamal(C_uBPrime)
		l4 := curve.Add(
			curve.ScalarMul(G, proof.Z_bar_r_B),
			curve.ScalarMul(C_uBPrimeNeg.CL, proof.Z_sk_uInv),
		)
		r4 := curve.Add(
			proof.A_T_uBC_uBRPrimeInv,
			curve.ScalarMul(
				curve.Add(
					proof.T_uB,
					C_uBPrimeNeg.CR,
				),
				c,
			),
		)
		if !l4.Equal(r4) {
			log.Println("[Verify AddLiquidityProof] l4 != r4")
			return false, nil
		}
		// fee part
		C_feePrimeNeg := &ElGamalEnc{
			CL: new(Point).Set(C_uAPrimeNeg.CL),
			CR: new(Point).Set(C_uAPrimeNeg.CR),
		}
		l5 := curve.Add(
			curve.ScalarMul(G, proof.Z_bar_r_fee),
			curve.ScalarMul(C_feePrimeNeg.CL, proof.Z_sk_uInv),
		)
		r5 := curve.Add(
			proof.A_T_feeC_feeRPrimeInv,
			curve.ScalarMul(
				curve.Add(
					proof.T_fee,
					C_feePrimeNeg.CR,
				),
				c,
			),
		)
		if !l5.Equal(r5) {
			log.Println("[Verify AddLiquidityProof] l5 != r5")
			return false, nil
		}
	} else if proof.GasFeeAssetId == proof.AssetBId {
		if !equalEnc(proof.C_uB, proof.C_fee) || !proof.A_T_uBC_uBRPrimeInv.Equal(proof.A_T_feeC_feeRPrimeInv) {
			log.Println("[Verify AddLiquidityProof] invalid params")
			return false, errors.New("[Verify AddLiquidityProof] invalid params")
		}
		// A part
		C_uAPrime, err = twistedElgamal.EncAdd(proof.C_uA, proof.C_uA_Delta)
		if err != nil {
			return false, err
		}
		C_uAPrimeNeg = negElgamal(C_uAPrime)
		l3 := curve.Add(
			curve.ScalarMul(G, proof.Z_bar_r_A),
			curve.ScalarMul(C_uAPrimeNeg.CL, proof.Z_sk_uInv),
		)
		r3 := curve.Add(
			proof.A_T_uAC_uARPrimeInv,
			curve.ScalarMul(
				curve.Add(
					proof.T_uA,
					C_uAPrimeNeg.CR,
				),
				c,
			),
		)
		if !l3.Equal(r3) {
			log.Println("[Verify AddLiquidityProof] l3 != r3")
			return false, nil
		}
		// B part
		C_uBPrime, err = twistedElgamal.EncAdd(proof.C_uB, proof.C_uB_Delta)
		if err != nil {
			return false, err
		}
		C_uBPrime.CR = curve.Add(C_uBPrime.CR, curve.ScalarMul(H, big.NewInt(-int64(proof.GasFee))))
		C_uBPrimeNeg = negElgamal(C_uBPrime)
		l4 := curve.Add(
			curve.ScalarMul(G, proof.Z_bar_r_B),
			curve.ScalarMul(C_uBPrimeNeg.CL, proof.Z_sk_uInv),
		)
		r4 := curve.Add(
			proof.A_T_uBC_uBRPrimeInv,
			curve.ScalarMul(
				curve.Add(
					proof.T_uB,
					C_uBPrimeNeg.CR,
				),
				c,
			),
		)
		if !l4.Equal(r4) {
			log.Println("[Verify AddLiquidityProof] l4 != r4")
			return false, nil
		}
		// fee part
		C_feePrimeNeg := &ElGamalEnc{
			CL: new(Point).Set(C_uBPrimeNeg.CL),
			CR: new(Point).Set(C_uBPrimeNeg.CR),
		}
		l5 := curve.Add(
			curve.ScalarMul(G, proof.Z_bar_r_fee),
			curve.ScalarMul(C_feePrimeNeg.CL, proof.Z_sk_uInv),
		)
		r5 := curve.Add(
			proof.A_T_feeC_feeRPrimeInv,
			curve.ScalarMul(
				curve.Add(
					proof.T_fee,
					C_feePrimeNeg.CR,
				),
				c,
			),
		)
		if !l5.Equal(r5) {
			log.Println("[Verify AddLiquidityProof] l5 != r5")
			return false, nil
		}
	} else {
		// A part
		C_uAPrime, err = twistedElgamal.EncAdd(proof.C_uA, proof.C_uA_Delta)
		if err != nil {
			return false, err
		}
		C_uAPrimeNeg = negElgamal(C_uAPrime)
		l3 := curve.Add(
			curve.ScalarMul(G, proof.Z_bar_r_A),
			curve.ScalarMul(C_uAPrimeNeg.CL, proof.Z_sk_uInv),
		)
		r3 := curve.Add(
			proof.A_T_uAC_uARPrimeInv,
			curve.ScalarMul(
				curve.Add(
					proof.T_uA,
					C_uAPrimeNeg.CR,
				),
				c,
			),
		)
		if !l3.Equal(r3) {
			log.Println("[Verify AddLiquidityProof] l3 != r3")
			return false, nil
		}
		// B paart
		C_uBPrime, err = twistedElgamal.EncAdd(proof.C_uB, proof.C_uB_Delta)
		if err != nil {
			return false, err
		}
		C_uBPrimeNeg = negElgamal(C_uBPrime)
		l4 := curve.Add(
			curve.ScalarMul(G, proof.Z_bar_r_B),
			curve.ScalarMul(C_uBPrimeNeg.CL, proof.Z_sk_uInv),
		)
		r4 := curve.Add(
			proof.A_T_uBC_uBRPrimeInv,
			curve.ScalarMul(
				curve.Add(
					proof.T_uB,
					C_uBPrimeNeg.CR,
				),
				c,
			),
		)
		if !l4.Equal(r4) {
			log.Println("[Verify AddLiquidityProof] l4 != r4")
			return false, nil
		}
		// fee part
		C_feeRPrime := curve.Add(proof.C_fee.CR, curve.ScalarMul(H, big.NewInt(-int64(proof.GasFee))))
		C_feePrime := &ElGamalEnc{
			CL: proof.C_fee.CL,
			CR: C_feeRPrime,
		}
		if err != nil {
			return false, err
		}
		C_feePrimeNeg := negElgamal(C_feePrime)
		l5 := curve.Add(
			curve.ScalarMul(G, proof.Z_bar_r_fee),
			curve.ScalarMul(C_feePrimeNeg.CL, proof.Z_sk_uInv),
		)
		r5 := curve.Add(
			proof.A_T_feeC_feeRPrimeInv,
			curve.ScalarMul(
				curve.Add(
					proof.T_fee,
					C_feePrimeNeg.CR,
				),
				c,
			),
		)
		if !l5.Equal(r5) {
			log.Println("[Verify AddLiquidityProof] l5 != r5")
			return false, nil
		}
	}
	// verify range proof
	var (
		addLiquidityRangeProofCount = 3
		rangeChan                   = make(chan int, addLiquidityRangeProofCount)
	)
	go verifyCtRangeRoutine(proof.ARangeProof, rangeChan)
	go verifyCtRangeRoutine(proof.BRangeProof, rangeChan)
	go verifyCtRangeRoutine(proof.GasFeePrimeRangeProof, rangeChan)
	for i := 0; i < addLiquidityRangeProofCount; i++ {
		val := <-rangeChan
		if val == ErrCode {
			log.Println("[Verify AddLiquidityProof] invalid range proof")
			return false, nil
		}
	}
	return true, nil
}

func verifyAddLiquidityParams(proof *AddLiquidityProof) (res bool, err error) {
	var (
		B_A_Delta, B_B_Delta int64
	)
	B_A_Delta = -int64(proof.B_A_Delta)
	B_B_Delta = -int64(proof.B_B_Delta)
	// C_uA_Delta
	C_uA_DeltaCL := curve.ScalarMul(proof.Pk_u, proof.R_DeltaA)
	C_uA_DeltaCRL := curve.ScalarBaseMul(proof.R_DeltaA)
	C_uA_DeltaCRR := curve.ScalarMul(H, big.NewInt(B_A_Delta))
	C_uA_DeltaCR := curve.Add(C_uA_DeltaCRL, C_uA_DeltaCRR)
	C_uA_Delta := &ElGamalEnc{
		CL: C_uA_DeltaCL,
		CR: C_uA_DeltaCR,
	}
	// C_uB_Delta
	C_uB_DeltaCL := curve.ScalarMul(proof.Pk_u, proof.R_DeltaB)
	C_uB_DeltaCRL := curve.ScalarBaseMul(proof.R_DeltaB)
	C_uB_DeltaCRR := curve.ScalarMul(H, big.NewInt(B_B_Delta))
	C_uB_DeltaCR := curve.Add(C_uB_DeltaCRL, C_uB_DeltaCRR)
	C_uB_Delta := &ElGamalEnc{
		CL: C_uB_DeltaCL,
		CR: C_uB_DeltaCR,
	}
	LC_poolA_Delta := &ElGamalEnc{
		CL: curve.ScalarMul(proof.Pk_pool, proof.R_DeltaA),
		CR: curve.Add(C_uA_DeltaCRL, curve.Neg(C_uA_DeltaCRR)),
	}
	LC_poolB_Delta := &ElGamalEnc{
		CL: curve.ScalarMul(proof.Pk_pool, proof.R_DeltaB),
		CR: curve.Add(C_uB_DeltaCRL, curve.Neg(C_uB_DeltaCRR)),
	}
	if !equalEnc(C_uA_Delta, proof.C_uA_Delta) || !equalEnc(C_uB_Delta, proof.C_uB_Delta) ||
		!equalEnc(LC_poolA_Delta, proof.LC_poolA_Delta) || !equalEnc(LC_poolB_Delta, proof.LC_poolB_Delta) {
		log.Println("[verifyAddLiquidityParams] invalid balance enc")
		return false, nil
	}
	// verify LP
	Delta_LPCheck := uint64(math.Floor(math.Sqrt(float64(proof.B_A_Delta * proof.B_B_Delta))))
	if Delta_LPCheck != proof.Delta_LP {
		log.Println("[verifyAddLiquidityParams] invalid LP")
		return false, nil
	}
	// verify AMM info & DAO balance info
	l := ffmath.Multiply(big.NewInt(int64(proof.B_poolB)), big.NewInt(int64(proof.B_A_Delta)))
	r := ffmath.Multiply(big.NewInt(int64(proof.B_poolA)), big.NewInt(int64(proof.B_B_Delta)))
	maxPool := int64(math.Floor(math.Max(float64(proof.B_poolA), float64(proof.B_poolB))))
	if new(big.Int).Abs(ffmath.Sub(l, r)).Cmp(big.NewInt(maxPool)) > 0 {
		log.Println("[verifyAddLiquidityParams] invalid liquidity rate")
		return false, nil
	}
	return true, nil
}

func (proof *AddLiquidityProof) AddPoolInfo(b_pool_A, b_pool_B uint64) (err error) {
	if !validUint64(b_pool_A) || !validUint64(b_pool_B) {
		log.Println("[AddPoolInfo] invalid params")
		return errors.New("[AddPoolInfo] invalid params")
	}
	proof.B_poolA = b_pool_A
	proof.B_poolB = b_pool_B
	return nil
}