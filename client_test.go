// Copyright (c) The Noxide Project Authors
// SPDX-License-Identifier: BSD-3-Clause

package memc

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/shoenig/test/must"
)

func Test_SetDialTimeout(t *testing.T) {
	t.Parallel()

	c := New(nil, SetDialTimeout(4*time.Second))
	must.Eq(t, 4*time.Second, c.timeout)
}

func Test_SetDefaultTTL(t *testing.T) {
	t.Parallel()

	c := New(nil, SetDefaultTTL(2*time.Hour))
	must.Eq(t, 2*time.Hour, c.expiration)
}

func Test_seconds(t *testing.T) {
	t.Parallel()

	t.Run("zero", func(t *testing.T) {
		s, err := seconds(0)
		must.NoError(t, err)
		must.Zero(t, s)
	})

	t.Run("millis", func(t *testing.T) {
		_, err := seconds(250 * time.Millisecond)
		must.ErrorIs(t, err, ErrExpiration)
	})

	t.Run("seconds", func(t *testing.T) {
		s, err := seconds(4 * time.Second)
		must.NoError(t, err)
		must.Eq(t, 4, s)
	})

	t.Run("month", func(t *testing.T) {
		ttl := 30 * 24 * time.Hour
		fix := ttl - (1 * time.Second)
		s, err := seconds(fix)
		must.NoError(t, err)
		must.Eq(t, 2591999, s)

		// TODO support for 1+ month values
	})
}

func Test_check(t *testing.T) {
	t.Parallel()

	t.Run("empty", func(t *testing.T) {
		must.ErrorIs(t, check(""), ErrKeyNotValid)
	})

	t.Run("normal", func(t *testing.T) {
		must.NoError(t, check("normal"))
	})

	t.Run("max", func(t *testing.T) {
		s := strings.Repeat("a", 250)
		must.NoError(t, check(s))
	})

	t.Run("long", func(t *testing.T) {
		s := strings.Repeat("a", 251)
		must.ErrorIs(t, check(s), ErrKeyNotValid)
	})

	t.Run("space", func(t *testing.T) {
		s := "abc 123"
		must.ErrorIs(t, check(s), ErrKeyNotValid)
	})

	t.Run("tab", func(t *testing.T) {
		s := "abc\t123"
		must.ErrorIs(t, check(s), ErrKeyNotValid)
	})
}

type person struct {
	Name string
	Age  int
}

func Test_encode(t *testing.T) {
	t.Parallel()

	t.Run("[]byte", func(t *testing.T) {
		b, err := encode([]byte{2, 4, 6, 8})
		must.NoError(t, err)
		must.SliceLen(t, 4, b)
	})

	t.Run("string", func(t *testing.T) {
		b, err := encode("foobar")
		must.NoError(t, err)
		must.SliceLen(t, 6, b)
	})

	t.Run("int8", func(t *testing.T) {
		var i int8 = 3
		b, err := encode(i)
		must.NoError(t, err)
		must.SliceLen(t, 1, b)
	})

	t.Run("uint8", func(t *testing.T) {
		var i uint8 = 3
		b, err := encode(i)
		must.NoError(t, err)
		must.SliceLen(t, 1, b)
	})

	t.Run("int16", func(t *testing.T) {
		var i int16 = math.MaxInt16
		b, err := encode(i)
		must.NoError(t, err)
		must.SliceLen(t, 2, b)
	})

	t.Run("uint16", func(t *testing.T) {
		var i uint16 = math.MaxUint16
		b, err := encode(i)
		must.NoError(t, err)
		must.SliceLen(t, 2, b)
	})

	t.Run("int32", func(t *testing.T) {
		var i int32 = math.MaxInt32
		b, err := encode(i)
		must.NoError(t, err)
		must.SliceLen(t, 4, b)
	})

	t.Run("uint32", func(t *testing.T) {
		var i uint32 = math.MaxUint32
		b, err := encode(i)
		must.NoError(t, err)
		must.SliceLen(t, 4, b)
	})

	t.Run("int64", func(t *testing.T) {
		var i int64 = math.MaxInt64
		b, err := encode(i)
		must.NoError(t, err)
		must.SliceLen(t, 8, b)
	})

	t.Run("uint64", func(t *testing.T) {
		var i uint64 = math.MaxUint64
		b, err := encode(i)
		must.NoError(t, err)
		must.SliceLen(t, 8, b)
	})

	t.Run("int", func(t *testing.T) {
		var i = math.MaxInt
		b, err := encode(i)
		must.NoError(t, err)
		must.SliceLen(t, 8, b)
	})

	t.Run("uint", func(t *testing.T) {
		var i uint = math.MaxUint
		b, err := encode(i)
		must.NoError(t, err)
		must.SliceLen(t, 8, b)
	})

	t.Run("struct", func(t *testing.T) {
		p := &person{
			Name: "bob",
			Age:  32,
		}
		b, err := encode(p)
		must.NoError(t, err)
		must.SliceLen(t, 48, b) // sure
	})
}

func Test_decode(t *testing.T) {
	t.Parallel()

	t.Run("[]byte", func(t *testing.T) {
		result, err := decode[[]byte]([]byte{1, 2})
		must.NoError(t, err)
		must.Eq(t, []byte{1, 2}, result)
	})

	t.Run("string", func(t *testing.T) {
		s := []byte("hello")
		result, err := decode[string](s)
		must.NoError(t, err)
		must.Eq(t, "hello", result)
	})

	t.Run("int8", func(t *testing.T) {
		result, err := decode[int8]([]byte{0xfe}) // little endian
		must.NoError(t, err)
		must.Eq(t, -2, result) // 2's compliment
	})

	t.Run("uint8", func(t *testing.T) {
		result, err := decode[uint8]([]byte{0xff})
		must.NoError(t, err)
		must.Eq(t, math.MaxUint8, result)
	})

	t.Run("int16", func(t *testing.T) {
		result, err := decode[int16]([]byte{0xfe, 0xff}) // little endian
		must.NoError(t, err)
		must.Eq(t, -2, result) // 2's compliment
	})

	t.Run("uint16", func(t *testing.T) {
		result, err := decode[uint16]([]byte{0xff, 0xff})
		must.NoError(t, err)
		must.Eq(t, math.MaxUint16, result)
	})

	t.Run("int32", func(t *testing.T) {
		result, err := decode[int32]([]byte{0xfe, 0xff, 0xff, 0xff}) // little endian
		must.NoError(t, err)
		must.Eq(t, -2, result) // 2's compliment
	})

	t.Run("uint32", func(t *testing.T) {
		result, err := decode[uint32]([]byte{0xff, 0xff, 0xff, 0xff})
		must.NoError(t, err)
		must.Eq(t, math.MaxUint32, result)
	})

	t.Run("int64", func(t *testing.T) {
		result, err := decode[int64]([]byte{0xfe, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) // little endian
		must.NoError(t, err)
		must.Eq(t, -2, result) // 2's compliment
	})

	t.Run("uint64", func(t *testing.T) {
		result, err := decode[uint64]([]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
		must.NoError(t, err)
		must.Eq(t, math.MaxUint64, result)
	})

	t.Run("int", func(t *testing.T) {
		result, err := decode[int]([]byte{0xfe, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) // little endian
		must.NoError(t, err)
		must.Eq(t, -2, result) // 2's compliment
	})

	t.Run("uint", func(t *testing.T) {
		result, err := decode[uint]([]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
		must.NoError(t, err)
		must.Eq(t, math.MaxUint, result)
	})

	t.Run("struct pointer", func(t *testing.T) {
		input, ierr := encode(&person{
			Name: "bob",
			Age:  32,
		})
		must.NoError(t, ierr)
		must.NotNil(t, input)

		result, err := decode[*person](input)
		must.NoError(t, err)
		must.Eq(t, &person{
			Name: "bob",
			Age:  32,
		}, result)
	})

	t.Run("struct value", func(t *testing.T) {
		input, ierr := encode(person{
			Name: "alice",
			Age:  30,
		})
		must.NoError(t, ierr)
		must.NotNil(t, input)

		result, err := decode[person](input)
		must.NoError(t, err)
		must.Eq(t, person{
			Name: "alice",
			Age:  30,
		}, result)
	})
}
