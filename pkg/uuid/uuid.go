// Copyright btrfscue authors
// SPDX-License-Identifier: BSD-2-Clause

// Custom data type for UUIDs

package uuid

import (
	"encoding/hex"
	"fmt"
	"strings"
)

const UUIDSize = 16

type UUID [UUIDSize]byte

var (
	uuidAllZero = UUID{}
	uuidAllFs   = UUID{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
)

func (u UUID) IsZero() bool {
	return u == uuidAllZero
}

func (u UUID) IsAllFs() bool {
	return u == uuidAllFs
}

func (u UUID) String() string {
	const hexChars = "0123456789abcdef"
	buf := [UUIDSize*2 + 4]byte{}
	p := 0
	for i, v := range u {
		buf[p] = hexChars[v>>4]
		p++
		buf[p] = hexChars[v&0xF]
		if i == 3 || i == 5 || i == 7 || i == 9 {
			p++
			buf[p] = '-'
		}
		p++
	}
	return string(buf[:])
}

func (u *UUID) Set(value string) error {
	b, err := hex.DecodeString(strings.Replace(value, "-", "", -1))
	if err != nil {
		return err
	}
	if len(b) != UUIDSize {
		return fmt.Errorf("expected %d hex characters plus optional dashes",
			UUIDSize*2)
	}
	copy(u[:], b)
	return nil
}

func (u UUID) Type() string { return "string" }

func New(value string) (UUID, error) {
	u := UUID{}
	err := u.Set(value)
	return u, err
}
