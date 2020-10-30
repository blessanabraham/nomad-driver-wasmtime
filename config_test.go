package main

import (
	"github.com/bytecodealliance/wasmtime-go"
	"testing"

	"github.com/hashicorp/nomad/helper/pluginutils/hclutils"
	"github.com/stretchr/testify/require"
)

func TestConfig_ParseHCL(t *testing.T) {
	cases := []struct {
		name string

		input    string
		expected *TaskConfig
	}{
		{
			"default strategy",
			`config {
				file = "add.wasm"
			}`,
			&TaskConfig{
				Compiler: WasmTimeCompiler{
					Strategy: "auto",
					CraneLiftOptions: CraneLiftOptions{
						OptLevel: wasmtime.OptLevelSpeed,
					},
				},
				Profiler: "none",
			},
		},
		{
			"cranelift defaults",
			`config {
				file = "add.wasm",
				compiler {
					strategy = "cranelift",
				},
			}`,
			&TaskConfig{
				Compiler: WasmTimeCompiler{
					Strategy: "cranelift",
					CraneLiftOptions: CraneLiftOptions{
						DebugVerifier:       false,
						OptLevel:            wasmtime.OptLevelSpeed,
						NANCanonicalization: false,
					},
				},
				Profiler: "none",
			},
		},
	}

	parser := hclutils.NewConfigParser(taskConfigSpec)
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			var tc *TaskConfig

			parser.ParseHCL(t, c.input, &tc)

			require.EqualValues(t, c.expected, tc)
		})
	}
}
