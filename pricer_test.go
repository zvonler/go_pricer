package main

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBadInput(t *testing.T) {
	pricer := NewPricer(1)

	// Empty line
	_, res := pricer.HandleLine("")
	require.Equal(t, "", res)

	// Bad time
	_, res = pricer.HandleLine("not_a_time")
	require.Equal(t, "", res)

	testLine := func(tm uint, input string, expected string) {
		res_tm, res := pricer.HandleLine(strconv.FormatUint(uint64(tm), 10) + " " + input)
		require.Equal(t, tm, res_tm)
		require.Equalf(t, expected, res, "tm %v", tm)
	}

	testLine(0, "", "")
	testLine(1, "A foo bad_side 12.34 100", "")
	testLine(2, "A bar B bad_price 100", "")
	testLine(3, "A baz B 12.34 bad_size", "")
	testLine(4, "R foo bad_size", "")
	testLine(5, "R unknown_oid 100", "")
}

func TestExample(t *testing.T) {
	pricer := NewPricer(200)

	testLine := func(tm uint, input string, expected string) {
		res_tm, res := pricer.HandleLine(strconv.FormatUint(uint64(tm), 10) + " " + input)
		require.Equal(t, tm, res_tm)
		require.Equalf(t, expected, res, "tm %v", tm)
	}

	testLine(28800538, "A b S 44.26 100", "")
	testLine(28800562, "A c B 44.10 100", "")
	testLine(28800744, "R b 100", "")
	testLine(28800758, "A d B 44.18 157", "S 8832.56")
	testLine(28800773, "A e S 44.38 100", "")
	testLine(28800796, "R d 157", "S NA")
	testLine(28800812, "A f B 44.18 157", "S 8832.56")
	testLine(28800974, "A g S 44.27 100", "B 8865.00")
	testLine(28800975, "R e 100", "B NA")
	testLine(28812071, "R f 100", "S NA")
	testLine(28813129, "A h B 43.68 50", "S 8806.50")
	testLine(28813300, "R f 57", "S NA")
	testLine(28813830, "A i S 44.18 100", "B 8845.00")
	testLine(28814087, "A j S 44.18 1000", "B 8836.00")
	testLine(28814834, "R c 100", "")
	testLine(28814864, "A k B 44.09 100", "")
	testLine(28815774, "R k 100", "")
	testLine(28815804, "A l B 44.07 175", "S 8804.25")
	testLine(28815937, "R j 1000", "B 8845.00")
	testLine(28816245, "A m S 44.22 100", "B 8840.00")
}
