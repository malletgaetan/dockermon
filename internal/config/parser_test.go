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

	for version, _ := range configVersion {
		f.Add(corpusBytes, version)
	}

	f.Fuzz(func(t *testing.T, data []byte, version string) {
		scanner := bufio.NewScanner(bytes.NewReader(data))
		config, err := ParseConfig(scanner, version)

		if err != nil {
			if config != nil {
				t.Errorf("got non-nil config with error: %v", err)
			}
			return
		}

		if config == nil {
			t.Error("got nil config without error")
		}
		if config.version != version {
			t.Errorf("version mismatch: got %s, expected %s", config.version, version)
		}
	})
}
