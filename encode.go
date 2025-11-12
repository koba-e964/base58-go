package base58

import (
	"crypto/subtle"
	"math/big"
	"math/bits"
	"unsafe"
)

// Using the idea described in https://github.com/btcsuite/btcd/blob/13152b35e191385a874294a9dbc902e48b1d71b0/btcutil/base58/base58.go#L34-L49
var (
	radix10  = new(big.Int).Exp(big.NewInt(58), big.NewInt(10), nil) // 58^10 < 2^64
	alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")
)

// VartimeEncode encodes a byte slice into a base58 string with length resultLength.
//
// This function does not have a constant-time guarantee.
func VartimeEncode(a []byte, resultLength int) string {
	tmp := big.NewInt(0)
	tmp.SetBytes(a)
	result := make([]byte, resultLength)
	for i := 0; i < resultLength; i += 10 {
		var remainder big.Int
		tmp.DivMod(tmp, radix10, &remainder)
		remainder64 := remainder.Uint64()
		for j := 0; j < 10; j++ {
			if i+j < resultLength {
				rem58 := remainder64 % 58
				remainder64 /= 58
				result[resultLength-1-i-j] = alphabet[int(rem58)]
			}
		}
	}
	return string(result)
}

// Encode encodes a byte slice into a base58 string with length resultLength.
//
// This function runs in constant time.
func Encode(a []byte, resultLength int) string {
	aLen := len(a)
	tmp := make([]uint32, (aLen+3)/4)
	for i := 0; i < aLen; i++ {
		tmp[len(tmp)-1-i/4] |= uint32(a[aLen-1-i]) << (8 * (i % 4))
	}
	result := make([]byte, resultLength)
	// log(58)/log(2) > 5.857 > 29/5, so every 5 letters we can delete 29 bits
	deletedBits := 0
	for i := 0; i < resultLength; i += 5 {
		rems := div58(tmp[min(len(tmp), deletedBits/32):])
		conv := func(remainder int) byte {
			char := '1' + remainder                                                                              // [0,9): '1'..'9'
			char = subtle.ConstantTimeSelect(subtle.ConstantTimeLessOrEq(9, remainder), 'A'+remainder-9, char)   // [9,17): 'A'..'H'
			char = subtle.ConstantTimeSelect(subtle.ConstantTimeLessOrEq(17, remainder), 'J'+remainder-17, char) // [17,22): 'J'..'N'
			char = subtle.ConstantTimeSelect(subtle.ConstantTimeLessOrEq(22, remainder), 'P'+remainder-22, char) // [22,33): 'P'..'Z'
			char = subtle.ConstantTimeSelect(subtle.ConstantTimeLessOrEq(33, remainder), 'a'+remainder-33, char) // [33,44): 'a'..'k'
			char = subtle.ConstantTimeSelect(subtle.ConstantTimeLessOrEq(44, remainder), 'm'+remainder-44, char) // [44,58): 'm'..'z'
			return byte(char)
		}
		result[resultLength-1-i] = conv(rems[0])
		for j := 1; j < 5; j++ {
			if i+j < resultLength {
				result[resultLength-1-i-j] = conv(rems[j])
			}
		}
		deletedBits += 29
	}
	return unsafe.String(unsafe.SliceData(result), len(result))
}

// constantTimeGeqUint64 returns 1 if a >= b, 0 otherwise, in constant time.
func constantTimeGeqUint64(a, b uint64) int {
	// Split into high and low 32 bits to safely use subtle functions
	// which take int parameters that may be 32-bit on some platforms.
	aHi := uint32(a >> 32)
	aLo := uint32(a)
	bHi := uint32(b >> 32)
	bLo := uint32(b)

	// a >= b if: aHi > bHi, OR (aHi == bHi AND aLo >= bLo)
	// Note: uint32 always fits in int (int is at least 32 bits in Go)
	hiGreater := subtle.ConstantTimeLessOrEq(int(bHi), int(aHi)) & ^subtle.ConstantTimeEq(int32(aHi), int32(bHi))
	hiEqual := subtle.ConstantTimeEq(int32(aHi), int32(bHi))
	loGeq := subtle.ConstantTimeLessOrEq(int(bLo), int(aLo))

	return hiGreater | (hiEqual & loGeq)
}

func div58(a []uint32) [5]int {
	// Using the idea described in https://github.com/btcsuite/btcd/blob/13152b35e191385a874294a9dbc902e48b1d71b0/btcutil/base58/base58.go#L34-L49
	// Using Barrett Reduction for constant-time division (https://kyberslash.cr.yp.to/)
	const d = 58 * 58 * 58 * 58 * 58 // 656356768
	// Barrett reduction constants for division by d
	const mBarrett64 = 28104751825 // floor(2^64 / 656356768)

	var carry uint64
	for i := 0; i < len(a); i++ {
		tmp := carry<<32 | uint64(a[i])
		// Barrett reduction: q ≈ (tmp * m) >> 64
		// For 64-bit tmp, we need to compute the high 64 bits of tmp * mBarrett64
		q, _ := bits.Mul64(tmp, mBarrett64)
		_, qd := bits.Mul64(q, d)
		r := tmp - qd
		// Correction step (constant-time)
		correction := uint64(subtle.ConstantTimeSelect(constantTimeGeqUint64(r, d), 1, 0))
		correctionD := -correction & d
		q += correction
		r -= correctionD
		a[i] = uint32(q)
		carry = r
	}

	// Barrett reduction constants for division by 58
	// m58 = floor(2^64 / 58) where k=64
	// Verification: 2^64 / 58 ≈ 318047311615681924.414, floor = 318047311615681924
	const mBarrett58 = 318047311615681924 // floor(2^64 / 58)

	var res [5]int
	for i := 0; i < 5; i++ {
		// Barrett reduction for division by 58
		q, _ := bits.Mul64(carry, mBarrett58)
		_, q58 := bits.Mul64(q, 58)
		r := carry - q58
		// Correction step (constant-time)
		correction := uint64(subtle.ConstantTimeSelect(constantTimeGeqUint64(r, 58), 1, 0))
		correction58 := -correction & 58
		q += correction
		r -= correction58
		res[i] = int(r)
		carry = q
	}
	return res
}
