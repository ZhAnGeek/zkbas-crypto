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

package transactions

import (
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/std/algebra/twistededwards"
	"github.com/consensys/gnark/std/hash/mimc"
	"zecrey-crypto/hash/bn254/zmimc"
	"zecrey-crypto/zecrey/circuit/bn254/std"
)

type BlockConstraints struct {
	Txs [35]TxConstraints
}

func (circuit BlockConstraints) Define(curveID ecc.ID, api API) error {
	// get edwards curve params
	params, err := twistededwards.NewEdCurve(curveID)
	if err != nil {
		return err
	}

	// mimc
	hFunc, err := mimc.NewMiMC(zmimc.SEED, curveID, api)
	if err != nil {
		return err
	}

	// TODO verify H: need to optimize
	H := Point{
		X: api.Constant(std.HX),
		Y: api.Constant(std.HY),
	}
	VerifyBlock(api, circuit, params, hFunc, H)

	return nil
}

func VerifyBlock(
	api API,
	block BlockConstraints,
	params twistededwards.EdCurve,
	hFunc MiMC,
	h Point,
) {
	for i := 0; i < len(block.Txs); i++ {
		VerifyTransaction(api, block.Txs[i], params, hFunc, h)
		hFunc.Reset()
	}
}

func SetBlockWitness(txs []TxConstraints, isEnabled bool) (witness BlockConstraints, err error) {
	for i := 0; i < len(txs); i++ {
		witness.Txs[i] = txs[i]
	}
	return witness, nil
}
