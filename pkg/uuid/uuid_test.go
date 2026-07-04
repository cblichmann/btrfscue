// Copyright btrfscue authors
// SPDX-License-Identifier: BSD-2-Clause

// Tests for UUID package

package uuid

import (
	"flag"
	"fmt"
	"testing"
)

var expected = UUID{0x7d, 0x00, 0x18, 0x96, 0x6b, 0x2d, 0x44, 0xc7, 0xbb, 0x8a,
	0xb5, 0xe8, 0x60, 0x1e, 0x8a, 0x7a}

func TestDefaults(t *testing.T) {
	if zeroes := (UUID{}); !zeroes.IsZero() {
		t.Fatalf("expected all zeroes, got: %s", zeroes)
	}
	if allfs := (UUID{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}); !allfs.IsAllFs() {
		t.Fatalf("expected all Fs, got: %s", allfs)
	}
}

func TestFromString(t *testing.T) {
	for _, v := range []string{
		"7d001896-6b2d-44c7-bb8a-b5e8601e8a7a",
		"7d0018966b2d44c7bb8ab5e8601e8a7a",
		"7d-00-18-96-6b-2d-44-c7bb8ab5e8601e8a7a",
		"7-d00-18966b2d44c7bb8ab5e8601e8a7a---",
		"7D0018966B2d44c7bb8AB5e8601e8a7a", // Mixed case
	} {
		if u, err := New(v); err != nil {
			t.Fatalf("expected no error, got: %v", err)
		} else if u != expected {
			t.Fatalf("expected %s, got: %s", expected, u)
		}
	}
}

func TestInvalid(t *testing.T) {
	u := UUID{}
	for _, v := range []string{
		"7d001896-6b2d-44c7-bb8a",
		"7d00zzzzzzzz44c7bb8ab5e8601e8a7a",
	} {
		if err := u.Set(v); err == nil {
			t.Fatalf("expected an error, got: %s", u)
		}
	}
}

func TestFormat(t *testing.T) {
	for _, v := range []struct {
		u UUID
		e string
	}{
		{UUID{}, "00000000-0000-0000-0000-000000000000"},
		{expected, "7d001896-6b2d-44c7-bb8a-b5e8601e8a7a"},
	} {
		if s := fmt.Sprintf("%v", v.u); s != v.e {
			t.Fatalf("expected %s, got: %s", v.e, s)
		}
	}
}

func TestFlag(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	u := UUID{}
	fs.Var(&u, "uuid", "")
	if err := fs.Parse([]string{"-uuid",
		"7d001896-6b2d-44c7-bb8a-b5e8601e8a7a"}); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if u != expected {
		t.Fatalf("expected %s, got: %s", expected, u)
	}

	if ty := u.Type(); ty != "string" {
		t.Fatalf("expected string, got: %s", ty)
	}
}
