package randtool

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"

	// "fmt"
	"math"
	"math/big"
)

func Uint16ToBytes(n uint16) []byte {
	return []byte{
		byte(n),
		byte(n >> 8),
	}
}

func BytesToUint32(buf []byte) uint32 {
	b_buf := bytes.NewBuffer(buf)
	var x uint32
	binary.Read(b_buf, binary.BigEndian, &x)
	return x
}

//将两个uint16的数字合并为一个uint32
func UnionUint16ToUint32(a uint16, b uint16) uint32 {
	a_buf := Uint16ToBytes(a)
	b_buf := Uint16ToBytes(b)

	a_buf = append(a_buf, b_buf[:]...)
	return BytesToUint32(a_buf)
}

// 生成区间[-m, n]的安全随机数
func RangeRand(min, max int64) int64 {
	if min > max {
		panic("the min is greater than max!")
	}

	if min < 0 {
		f64Min := math.Abs(float64(min))
		i64Min := int64(f64Min)
		result, _ := rand.Int(rand.Reader, big.NewInt(max+1+i64Min))

		return result.Int64() - i64Min
	} else {
		result, _ := rand.Int(rand.Reader, big.NewInt(max-min+1))
		return min + result.Int64()
	}
}
