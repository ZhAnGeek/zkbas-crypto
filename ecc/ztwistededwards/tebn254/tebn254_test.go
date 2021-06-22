package tebn254

import (
	"fmt"
	"math/big"
	"testing"
	"zecrey-crypto/ffmath"
)

func TestNeg(t *testing.T) {
	a := big.NewInt(3)
	b := big.NewInt(-3)
	c := ffmath.ModInverse(a, Order)
	d := ffmath.Mod(b, Order)
	A := ScalarBaseMul(a)
	B := ScalarBaseMul(b)
	C := ScalarMul(A, c)
	D := ScalarBaseMul(d)
	E := Add(A, D)
	fmt.Println(E.Equal(&O))
	fmt.Println(C.Equal(G))
	AB := Add(A, B)
	fmt.Println(A)
	fmt.Println(B)
	fmt.Println(AB)
	GNeg := Neg(G)
	GNeg2 := Add(GNeg, GNeg)
	GNeg3 := Add(GNeg2, GNeg)
	fmt.Println(GNeg3)
	ANeg := Neg(A)
	ANeg2 := ScalarMul(A, big.NewInt(-1))
	fmt.Println(A)
	fmt.Println(B)
	fmt.Println(ANeg)
	fmt.Println(ANeg2)
	C = Add(A, ANeg)
	C2 := Add(A, ANeg2)
	fmt.Println(C)
	fmt.Println(C2)
	fmt.Println(IsZero(C))
}

func TestMapToGroup(t *testing.T) {
	HTest, err := MapToGroup("zecreyHSeed")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(HTest)
	fmt.Println(H)
	fmt.Println(U)
}

func TestAdd(t *testing.T) {
	r1 := big.NewInt(3)
	r2 := big.NewInt(9)
	A1 := ScalarBaseMul(r1)
	Neutral := ZeroPoint()
	fmt.Println(Neutral.X.String())
	fmt.Println(Neutral.Y.String())
	A1Copy := Add(A1, Neutral)
	fmt.Println(A1)
	fmt.Println(A1Copy)
	A2 := ScalarBaseMul(r2)
	fmt.Println(A2)
	fmt.Println(A1)
	A2 = Add(A2, A1)
	fmt.Println(A2)
	fmt.Println(A1)
}

func TestAssign(t *testing.T) {
	A := ScalarBaseMul(big.NewInt(230928302))
	fmt.Println(A.X)
}
