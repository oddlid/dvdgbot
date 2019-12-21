package leet

import (
	"testing"
)

const (
	BC_TEST_JSON_FILE = "/tmp/bonusconfigs_test.json"
)

var (
	_testbcs = BonusConfigs{
		BonusConfig{
			Greeting:     "The ultimate goal!",
			SubString:    "1337",
			PrefixChar:   '0',
			UseStep:      true,
			StepPoints:   10,
			NoStepPoints: 0,
		},
		BonusConfig{
			Greeting:     `Hail Satan \m/`,
			SubString:    "666", // because, of course...
			PrefixChar:   '0',   // not used, as UseStep is false
			UseStep:      false, //
			StepPoints:   0,     // not used
			NoStepPoints: 18,    // 18 points, because 6+6+6 = 18
		},
	}
)

func TestBCWriteFile(t *testing.T) {
	err := _testbcs.saveFile(BC_TEST_JSON_FILE)
	if err != nil {
		t.Error(err)
	}
}

func TestBCReadFile(t *testing.T) {
	var bcs BonusConfigs
	err := bcs.loadFile(BC_TEST_JSON_FILE)
	if err != nil {
		t.Error(err)
	}

	for _, bc := range bcs {
		t.Logf("%#v", bc)
	}
}
