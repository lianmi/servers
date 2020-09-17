package ECDH

import (
	"crypto/rand"
	"io"
	"golang.org/x/crypto/curve25519"
)


func GetCurve25519KeypPair() (Aprivate, Apublic [32]byte)  {
	//产生随机数
	if _, err := io.ReadFull(rand.Reader, Aprivate[:]); err != nil {
		return
	}
	curve25519.ScalarBaseMult(&Apublic, &Aprivate)
	return
}

func GetCurve25519Key(private, public [32]byte) (Key [32]byte) {
	curve25519.ScalarMult(&Key, &private, &public)
	return
}
