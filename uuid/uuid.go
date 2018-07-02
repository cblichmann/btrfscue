/*
 * btrfscue version 0.5
 * Copyright (c)2011-2018 Christian Blichmann
 *
 * Custom data type for UUIDs
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are met:
 *     * Redistributions of source code must retain the above copyright
 *       notice, this list of conditions and the following disclaimer.
 *     * Redistributions in binary form must reproduce the above copyright
 *       notice, this list of conditions and the following disclaimer in the
 *       documentation and/or other materials provided with the distribution.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
 * AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
 * ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
 * LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
 * CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
 * SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
 * INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
 * CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
 * ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
 * POSSIBILITY OF SUCH DAMAGE.
 */

package uuid // import "blichmann.eu/code/btrfscue/uuid"

import (
	"encoding/hex"
	"fmt"
	"strings"
)

const UUIDSize = 16

type UUID [UUIDSize]byte

func (u UUID) IsZero() bool {
	return u == UUID{}
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
	if len(b) != UUIDSize {
		err = fmt.Errorf("expected %d hex characters plus optional dashes",
			UUIDSize*2)
	}
	if err != nil {
		return err
	}
	copy(u[:], b)
	return nil
}
