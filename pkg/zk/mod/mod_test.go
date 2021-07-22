package zkmod

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/cronokirby/safenum"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taurusgroup/cmp-ecdsa/pkg/hash"
	"github.com/taurusgroup/cmp-ecdsa/pkg/math/sample"
	"github.com/taurusgroup/cmp-ecdsa/pkg/zk"
)

func TestMod(t *testing.T) {
	p, q := zk.ProverPaillierSecret.P(), zk.ProverPaillierSecret.Q()
	sk := zk.ProverPaillierSecret
	public := Public{N: sk.PublicKey.N()}
	proof := NewProof(hash.New(), public, Private{
		P:   p,
		Q:   q,
		Phi: sk.Phi(),
	})
	out, err := proof.Marshal()
	require.NoError(t, err, "failed to marshal proof")
	proof2 := &Proof{}
	require.NoError(t, proof2.Unmarshal(out), "failed to unmarshal proof")
	out2, err := proof2.Marshal()
	require.NoError(t, err, "failed to marshal 2nd proof")
	proof3 := &Proof{}
	require.NoError(t, proof3.Unmarshal(out2), "failed to unmarshal 2nd proof")

	assert.True(t, proof3.Verify(hash.New(), public))

	proof.W = big.NewInt(0)
	for idx := range *proof.X {
		(*proof.X)[idx] = big.NewInt(0)
	}

	assert.False(t, proof.Verify(hash.New(), public), "proof should have failed")
}

func Test_set4thRoot(t *testing.T) {
	var p, q uint64 = 311, 331
	pMod := safenum.ModulusFromUint64(p)
	pHalf := new(safenum.Nat).SetUint64((p - 1) / 2)
	qMod := safenum.ModulusFromUint64(q)
	qHalf := new(safenum.Nat).SetUint64((q - 1) / 2)
	n := safenum.ModulusFromUint64(p * q)
	nBig := big.NewInt(int64(p * q))
	phi := new(safenum.Nat).SetUint64((p - 1) * (q - 1))
	y := new(safenum.Nat).SetUint64(502)
	wBig := sample.QNR(rand.Reader, nBig)
	w := new(safenum.Nat).SetBig(wBig, wBig.BitLen())

	a, b, x := makeQuadraticResidue(y, w, pHalf, qHalf, n, pMod, qMod)

	root := fourthRoot(x, phi, n)

	if b {
		y.ModMul(y, w, n)
	}
	if a {
		y.ModNeg(y, n)
	}

	assert.NotEqual(t, root, big.NewInt(1), "root cannot be 1")
	root.Exp(root, new(safenum.Nat).SetUint64(4), n)
	assert.True(t, root.Eq(y) == 1, "root^4 should be equal to y")
}
