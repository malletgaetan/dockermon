package config

import (
	"bufio"
	"bytes"
	"os"
	"testing"
)

func FuzzParseConfig(f *testing.F) {
	corpusBytes, err := os.ReadFile("../../configs/corpus.conf")
	if err != nil {
		f.Fatal(err)
	}

	f.Add(corpusBytes)

	f.Fuzz(func(t *testing.T, data []byte) {
		hints, _ := configVersion["1.47"]
		scanner := bufio.NewScanner(bytes.NewReader(data))
		config, err := ParseConfig(scanner, hints)

		if err != nil {
			if config != nil {
				t.Errorf("got non-nil config with error: %v", err)
			}
			return
		}

		if config == nil {
			t.Error("got nil config without error")
		}
	})
}
