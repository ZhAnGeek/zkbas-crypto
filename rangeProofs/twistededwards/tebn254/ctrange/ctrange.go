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

package ctrange

import (
	"bytes"
	"errors"
	"math/big"
	"zecrey-crypto/commitment/twistededwards/tebn254/pedersen"
	curve "zecrey-crypto/ecc/ztwistededwards/tebn254"
	"zecrey-crypto/ffmath"
	"zecrey-crypto/hash/bn254/zmimc"
	"zecrey-crypto/util"
)

var bitchan = make(chan int, RangeMaxBits)

func Prove(b int64, g, h *Point) (proof *RangeProof, err error) {
	// check params
	if b < 0 || !curve.IsInSubGroup(g) || !curve.IsInSubGroup(h) {
		return nil, errors.New("[ctrange Prove] err: invalid params")
	}
	// define new variable
	proof = new(RangeProof)
	proof.G = g
	proof.H = h
	// convert b into binary
	var (
		alphas [RangeMaxBits]*big.Int
		rs     [RangeMaxBits]*big.Int
		cs     [RangeMaxBits]*big.Int
		A_As   [RangeMaxBits]*Point
		buf    bytes.Buffer
	)
	bs, err := toBinary(b, RangeMaxBits)
	if err != nil {
		return nil, err
	}
	current := curve.H
	for i, bi := range bs {
		go phase1(i, bi, current, proof, &alphas, &rs, &cs, &A_As, g, h)
		current = curve.Add(current, current)
	}
	for i := 0; i < RangeMaxBits; i++ {
		index := <-bitchan
		if index == ErrCode {
			return nil, errors.New("[ctrange Prove] unable to construct the proof")
		}
	}
	for i := 0; i < RangeMaxBits; i++ {
		buf.Write(A_As[i].X.Marshal())
		buf.Write(A_As[i].Y.Marshal())
	}
	proof.C, err = util.HashToInt(buf, zmimc.Hmimc)
	if err != nil {
		return nil, err
	}
	proof.C = ffmath.Mod(proof.C, Q)
	twoExp := int64(1)
	for i, bi := range bs {
		go phase2(i, bi, twoExp, proof, &rs, &cs, &alphas, g, h)
		twoExp += twoExp
	}
	for i := 0; i < RangeMaxBits; i++ {
		index := <-bitchan
		if index == ErrCode {
			return nil, errors.New("[ctrange Prove] unable to construct the proof")
		}
	}
	proof.A = curve.ZeroPoint()
	for i := 0; i < RangeMaxBits; i++ {
		proof.A = curve.Add(proof.A, proof.As[i])
	}
	return proof, nil
}

func phase1(i int, bi int, htwoExp *Point, proof *RangeProof, alphas, rs, cs *[RangeMaxBits]*big.Int, A_As *[RangeMaxBits]*Point, g, h *Point) {
	var (
		err error
		buf bytes.Buffer
	)
	if bi == 0 {
		alphas[i] = curve.RandomValue()
		A_As[i] = curve.ScalarMul(g, alphas[i])
		bitchan <- i
	} else if bi == 1 {
		rs[i] = curve.RandomValue()
		proof.As[i] = curve.Add(curve.ScalarMul(g, rs[i]), htwoExp)
		if err != nil {
			bitchan <- ErrCode
		}
		alphas[i] = curve.RandomValue()
		galphai := curve.ScalarMul(g, alphas[i])
		buf.Write(galphai.X.Marshal())
		buf.Write(galphai.Y.Marshal())
		cs[i], err = util.HashToInt(buf, zmimc.Hmimc)
		if err != nil {
			bitchan <- ErrCode
		}
		A_As[i] = curve.ScalarMul(proof.As[i], cs[i])
		bitchan <- i
	} else {
		bitchan <- ErrCode
	}
}

func phase2(
	i int, bi int, twoExp int64, proof *RangeProof,
	rs *[RangeMaxBits]*big.Int, cs *[RangeMaxBits]*big.Int, alphas *[RangeMaxBits]*big.Int,
	g, h *Point,
) {
	var (
		err error
		tmp *Point
		buf bytes.Buffer
	)
	if bi == 0 {
		rs[i] = curve.RandomValue()
		tmp, err = pedersen.Commit(ffmath.Multiply(proof.C, big.NewInt(twoExp)), rs[i], h, g)
		if err != nil {
			bitchan <- ErrCode
		}
		buf.Write(tmp.X.Marshal())
		buf.Write(tmp.Y.Marshal())
		cs[i], err = util.HashToInt(buf, zmimc.Hmimc)
		if err != nil {
			bitchan <- ErrCode
		}
		proof.As[i] = curve.ScalarMul(g, ffmath.Multiply(alphas[i], ffmath.ModInverse(cs[i], Order)))
		proof.Zs[i] = ffmath.AddMod(rs[i], ffmath.Multiply(ffmath.Multiply(alphas[i], proof.C), ffmath.ModInverse(cs[i], Order)), Order)
		bitchan <- i
	} else if bi == 1 {
		proof.Zs[i] = ffmath.AddMod(alphas[i], ffmath.Multiply(proof.C, rs[i]), Order)
		bitchan <- i
	} else {
		bitchan <- ErrCode
	}
}

func (proof *RangeProof) Verify() (bool, error) {
	var (
		current *Point
		A_As    [RangeMaxBits]*Point
		A       *Point
		buf     bytes.Buffer
	)
	A = curve.ZeroPoint()
	current = curve.Neg(proof.H)
	for i := 0; i < RangeMaxBits; i++ {
		AihNeg := curve.Add(proof.As[i], current)
		AihNegNeg := curve.Neg(AihNeg)
		AihNegNeg = curve.ScalarMul(AihNegNeg, proof.C)
		com := curve.Add(curve.ScalarMul(proof.G, proof.Zs[i]), AihNegNeg)
		buf.Write(com.X.Marshal())
		buf.Write(com.Y.Marshal())
		ci, err := util.HashToInt(buf, zmimc.Hmimc)
		if err != nil {
			return false, err
		}
		A_As[i] = curve.ScalarMul(proof.As[i], ci)
		current = curve.Add(current, current)
		buf.Reset()
	}
	for _, A_Ai := range A_As {
		buf.Write(A_Ai.X.Marshal())
		buf.Write(A_Ai.Y.Marshal())
	}
	hatc, err := util.HashToInt(buf, zmimc.Hmimc)
	if err != nil {
		return false, err
	}
	hatc = ffmath.Mod(hatc, Q)
	if !ffmath.Equal(hatc, proof.C) {
		return false, nil
	}
	for _, Ai := range proof.As {
		A = curve.Add(A, Ai)
	}
	if !A.Equal(proof.A) {
		return false, nil
	}
	return true, nil
}
