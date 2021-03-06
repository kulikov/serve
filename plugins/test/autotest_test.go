package test

import (
	"fmt"
	"testing"

	"github.com/ghodss/yaml"

	"github.com/InnovaCo/serve/manifest"
	"github.com/InnovaCo/serve/utils"
)

func loadTestData(data string, plugin manifest.Plugin) error {
	if json, err := yaml.YAMLToJSON([]byte(data)); err != nil {
		return fmt.Errorf("Parser error: %v", err)
	} else {
		return plugin.Run(*manifest.LoadJSON(string(json)))
	}
}

type processorTestCase struct {
	in     string
	expect map[string]interface{}
}

func TestTestAutotest(t *testing.T) {
	runAllMultiCmdTests(t,
		map[string]processorTestCase{
			"simple": {
				in: `---
project: "test"
version: "0.0.0"
repo: "git@test.ru:test.git"
suite: "test-test"
environment: {}`,
				expect: map[string]interface{}{
					"cmdline": []string{"rm -rf autotest && git clone --depth 1 --single-branch --recursive git@test.ru:test.git autotest",
						"cd autotest/ && ./test.sh --project=test --version=0.0.0 --suite=test-test"},
				},
			},
		},
		TestAutotest{})
}

func runAllMultiCmdTests(t *testing.T, cases map[string]processorTestCase, plugin manifest.Plugin) {
	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			utils.RunCmdWithEnv = func(cmdline string, env map[string]string) error {
				for _, v := range test.expect["cmdline"].([]string) {
					if v == cmdline {
						return nil
					}
				}
				return fmt.Errorf("cmd: %v not found in %v", cmdline, test.expect["cmdline"].([]string))
			}

			if err := loadTestData(test.in, plugin); err != nil {
				t.Errorf("Error: %v", err)
				t.Fail()
			}
		})
	}
}
