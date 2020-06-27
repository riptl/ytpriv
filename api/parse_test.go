package api

import "testing"

func TestParseDuration(t *testing.T) {
	tests := []string{
		"PT1M9S",
		"PT1M09S",
	}

	results := []uint64{
		69,
		69,
	}

	for i := range tests {
		n, err := ParseDuration(tests[i])
		if n != results[i] {
			t.Errorf("%s => \n"+
				"Expected: %d, got: %d",
				tests[i], results[i], n,
			)
		} else if err != nil {
			t.Error(tests[i], err)
		}
	}

	fails := []string{
		"PS23M23S",
		"PT30",
		"PT30S",
		"PTXDM30S",
		"PT23MXDS",
		"PTMS",
	}

	for _, fail := range fails {
		if n, err := ParseDuration(fail); err == nil {
			t.Errorf("Didn't return an an error.\n"+
				"(%s => %d)", fail, n)
		}
	}
}

func TestExtractNumber(t *testing.T) {
	tests := []string{
		"123",
		"123,456",
	}

	results := []uint64{
		123,
		123456,
	}

	for i := range tests {
		n, err := ExtractNumber(tests[i])
		if n != results[i] {
			t.Errorf("%s => \n"+
				"Expected: %d, got: %d",
				tests[i], results[i], n,
			)
		} else if err != nil {
			t.Error(tests[i], err)
		}
	}
}
