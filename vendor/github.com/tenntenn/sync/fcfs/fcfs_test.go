package fcfs_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	. "github.com/tenntenn/sync/fcfs"
)

func TestGroup_Go(t *testing.T) {

	cases := []struct {
		name      string
		funcs     []func() (interface{}, error)
		value     interface{}
		expectErr bool
	}{
		{
			name:      "empty",
			expectErr: true,
		},
		{
			name: "one",
			funcs: []func() (interface{}, error){
				func() (interface{}, error) {
					return 100, nil
				},
			},
			value: 100,
		},
		{
			name: "two",
			funcs: []func() (interface{}, error){
				func() (interface{}, error) {
					time.Sleep(100 * time.Second)
					return 100, nil
				},
				func() (interface{}, error) {
					return 200, nil
				},
			},
			value: 200,
		},
		{
			name: "one-error",
			funcs: []func() (interface{}, error){
				func() (interface{}, error) {
					return 0, errors.New("error")
				},
				func() (interface{}, error) {
					return 200, nil
				},
			},
			value: 200,
		},
		{
			name: "all-error",
			funcs: []func() (interface{}, error){
				func() (interface{}, error) {
					return 0, errors.New("error")
				},
				func() (interface{}, error) {
					return 0, errors.New("error")
				},
			},
			expectErr: true,
		},
	}

	for i := range cases {
		c := cases[i]
		t.Run(c.name, func(t *testing.T) {
			var g Group
			for _, f := range c.funcs {
				g.Go(f)
			}
			v, err := g.Wait()
			switch {
			case err != nil && !c.expectErr:
				t.Error("unexpected error", err)
			case err == nil && c.expectErr:
				t.Error("expected error had not occured")
			}

			if !cmp.Equal(c.value, v) {
				t.Errorf("want %v\ngot %v", c.value, v)
			}
		})
	}
}
