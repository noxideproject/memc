// Copyright (c) The Noxide Project Authors
// SPDX-License-Identifier: BSD-3-Clause

package memc

import (
	"fmt"
	"net"
	"os/exec"
	"testing"
	"time"

	"github.com/shoenig/ignore"
	"github.com/shoenig/test/must"
	"github.com/shoenig/test/portal"
	"github.com/shoenig/test/skip"
	"github.com/shoenig/test/wait"
	"noxide.lol/go/xtc"
)

const (
	executable = "memcached"
)

type fatalTester struct{}

func (ft *fatalTester) Fatalf(msg string, args ...any) {
	s := fmt.Sprintf(msg, args...)
	panic(s)
}

var (
	fatal = new(fatalTester)
	ports = portal.New(fatal)
)

func launchTCP(t *testing.T, args []string) (string, func()) {
	// requires memcached executable on $PATH
	skip.CommandUnavailable(t, executable)

	port := ports.One()
	address := fmt.Sprintf("localhost:%d", port)
	args = append(args, "-l", address)

	ctx, cancel := xtc.Cancelable()
	cmd := exec.CommandContext(ctx, executable, args...)
	err := cmd.Start()
	must.NoError(t, err)

	// wait for memcached to be listening
	must.Wait(t, wait.InitialSuccess(
		wait.Timeout(3*time.Second),
		wait.Gap(200*time.Millisecond),
		wait.ErrorFunc(func() error {
			_, err := net.Dial("tcp", address)
			return err
		}),
	))

	return address, cancel
}

func TestE2E_SetGet_simple(t *testing.T) {
	t.Parallel()

	address, done := launchTCP(t, nil)
	t.Cleanup(done)

	c := New([]string{address})
	defer ignore.Close(c)

	t.Run("string", func(t *testing.T) {
		err := Set(c, "mystring", "myvalue")
		must.NoError(t, err)

		var v string
		v, err = Get[string](c, "mystring")
		must.NoError(t, err)
		must.Eq(t, "myvalue", v)
	})

	t.Run("[]byte", func(t *testing.T) {
		err := Set(c, "mybytes", []byte{2, 4, 6, 8})
		must.NoError(t, err)

		var v []byte
		v, err = Get[[]byte](c, "mybytes")
		must.NoError(t, err)
		must.Eq(t, []byte{2, 4, 6, 8}, v)
	})

	t.Run("int", func(t *testing.T) {
		err := Set(c, "myint", 998877)
		must.NoError(t, err)

		var v int
		v, err = Get[int](c, "myint")
		must.NoError(t, err)
		must.Eq(t, 998877, v)
	})

	t.Run("struct pointer", func(t *testing.T) {
		p := &person{Name: "Seth", Age: 34}
		err := Set(c, "myperson_p", p)
		must.NoError(t, err)

		var v *person
		v, err = Get[*person](c, "myperson_p")
		must.NoError(t, err)
		must.Eq(t, &person{Name: "Seth", Age: 34}, v)
	})

	t.Run("struct value", func(t *testing.T) {
		p := person{Name: "Seth", Age: 34}
		err := Set(c, "myperson_v", p)
		must.NoError(t, err)

		var v person
		v, err = Get[person](c, "myperson_v")
		must.NoError(t, err)
		must.Eq(t, person{Name: "Seth", Age: 34}, v)
	})
}

func Test_SetGet_expiration(t *testing.T) {
	t.Parallel()

	address, done := launchTCP(t, nil)
	t.Cleanup(done)

	c := New([]string{address})
	defer ignore.Close(c)

	t.Run("hour", func(t *testing.T) {
		err := Set(c, "mykey", "myvalue", TTL(1*time.Hour))
		must.NoError(t, err)
	})
}

func Test_Get_miss(t *testing.T) {
	t.Parallel()

	address, done := launchTCP(t, nil)
	t.Cleanup(done)

	c := New([]string{address})
	defer ignore.Close(c)

	_, err := Get[string](c, "missing")
	must.ErrorIs(t, err, ErrCacheMiss)
}
