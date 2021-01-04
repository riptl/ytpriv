package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeVideoID(t *testing.T) {
	type testCase struct {
		id  string
		num int64
	}
	cases := []testCase{
		{id: "YPiOWJDdChM", num: int64(6987491332903995923)},
		{id: "YLUE-m2B5xA", num: int64(6968481472051275536)},
		{id: "gq1uiinNM_Y", num: int64(-9030440138122120202)},
	}
	for _, c := range cases {
		num, err := decodeVideoID(c.id)
		require.NoError(t, err)
		assert.Equal(t, c.num, num)
		assert.Equal(t, c.id, encodeVideoID(num))
	}
}
