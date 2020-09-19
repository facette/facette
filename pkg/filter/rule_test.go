// Copyright (c) 2020, The Facette Authors
//
// Licensed under the terms of the BSD 3-Clause License; a copy of the license
// is available at: https://opensource.org/licenses/BSD-3-Clause

package filter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_FilterAction_UnmarshalText(t *testing.T) {
	for _, test := range []struct {
		input    []byte
		expected Action
		err      string
	}{
		{
			input:    []byte("discard"),
			expected: ActionDiscard,
		},
		{
			input:    []byte("relabel"),
			expected: ActionRelabel,
		},
		{
			input:    []byte("rewrite"),
			expected: ActionRewrite,
		},
		{
			input:    []byte("sieve"),
			expected: ActionSieve,
		},
		{
			input: []byte(""),
			err:   "invalid filter action",
		},
		{
			input: []byte("invalid"),
			err:   "invalid filter action: invalid",
		},
	} {
		var action Action

		err := action.UnmarshalText(test.input)
		if test.err == "" {
			assert.Nil(t, err)
			assert.Equal(t, test.expected, action)
		} else {
			assert.EqualError(t, err, test.err)
		}
	}
}

func Test_FilterPattern_MarshalText(t *testing.T) {
	b, err := Pattern{s: "^(abc)$"}.MarshalText()
	assert.Nil(t, err)
	assert.Equal(t, []byte("^(abc)$"), b)
}

func Test_FilterPattern_UnmarshalText(t *testing.T) {
	pattern := Pattern{}
	err := pattern.UnmarshalText([]byte("^(abc)$"))
	assert.Nil(t, err)
	assert.Equal(t, "^(abc)$", pattern.s)

	pattern = Pattern{}
	err = pattern.UnmarshalText([]byte("^(abc$"))
	assert.NotNil(t, err)
	assert.EqualError(t, err, "invalid filter pattern: missing closing ): `^(abc$`")
}
