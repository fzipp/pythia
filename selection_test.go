// Copyright 2013 Frederik Zipp.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"testing"
)

func TestNthIndexByte(t *testing.T) {
	s := []byte("xx.xxx.xxxx.xxxxx.xx.xxxxxxx.x")
	tests := []struct {
		s    []byte
		c    byte
		n    int
		want int
	}{
		{s, 'y', 1, -1},
		{s, '.', -1, -1},
		{s, '.', 0, -1},
		{s, '.', 1, 2},
		{s, '.', 2, 6},
		{s, '.', 3, 11},
		{s, '.', 4, 17},
		{s, '.', 5, 20},
		{s, '.', 6, 28},
		{s, '.', 7, -1},
	}
	for _, tt := range tests {
		if i := nthIndexByte(tt.s, tt.c, tt.n); i != tt.want {
			t.Errorf("nthIndexByte(%q, %q, %d) = %d, want %d", tt.s, tt.c, tt.n, i, tt.want)
		}
	}
}
