package main

import (
	"github.com/bytecodealliance/wasmtime-go"
	"github.com/hashicorp/nomad/plugins/base"
	"github.com/hashicorp/nomad/plugins/drivers"
	"github.com/hashicorp/nomad/plugins/shared/hclspec"
)

const (
	// pluginName is the name of the plugin
	// this is used for logging and (along with the version) for uniquely
	// identifying plugin binaries fingerprinted by the client
	pluginName = "wasmtime"

	// pluginVersion allows the client to identify and use newer versions of
	// an installed plugin
	pluginVersion = "v0.1.0"
)

var (
	// pluginInfo describes the plugin
	pluginInfo = &base.PluginInfoResponse{
		Type:              base.PluginTypeDriver,
		PluginApiVersions: []string{drivers.ApiVersion010},
		PluginVersion:     pluginVersion,
		Name:              pluginName,
	}

	// configSpec is the specification of the plugin's configuration
	// this is used to validate the configuration specified for the plugin
	// on the client.
	// this is not global, but can be specified on a per-client basis.
	configSpec = hclspec.NewObject(map[string]*hclspec.Spec{
		// The schema should be defined using HCL specs and it will be used to
		// validate the agent configuration provided by the user in the
		// `plugin` stanza (https://www.nomadproject.io/docs/configuration/plugin.html).
		//
		// For example, for the schema below a valid configuration would be:
		//
		//   plugin "wasmtime" {
		//     config {
		//       compiler = "cranelift"
		//		 profiler = "vtune"
		//     }
		//   }

	})

	// taskConfigSpec is the specification of the plugin's configuration for
	// a task
	// this is used to validated the configuration specified for the plugin
	// when a job is submitted.
	taskConfigSpec = hclspec.NewObject(map[string]*hclspec.Spec{
		// TODO: define plugin's task configuration schema
		//
		// The schema should be defined using HCL specs and it will be used to
		// validate the task configuration provided by the user when they
		// submit a job.
		//
		// For example, for the schema below a valid task would be:
		//   job "example" {
		//     group "example" {
		//       task "say-hi" {
		//         driver = "wasmtime"
		//         config {
		//			 file = "add.wasm"
		//         }
		//       }
		//     }
		//   }
		"file": hclspec.NewAttr("file", "string", true),
		"compiler": hclspec.NewDefault(
			hclspec.NewBlock("compiler", false, hclspec.NewObject(map[string]*hclspec.Spec{
				"strategy": hclspec.NewAttr("strategy", "string", false),
				"cranelift_options": hclspec.NewDefault(hclspec.NewBlock("cranelift_options", false, hclspec.NewObject(map[string]*hclspec.Spec{
					"debug_verifier": hclspec.NewDefault(
						hclspec.NewAttr("debug_verifier", "bool", false),
						hclspec.NewLiteral(`false`),
					),
					"optimize": hclspec.NewDefault(
						hclspec.NewAttr("optimize", "number", false),
						hclspec.NewLiteral(`1`),
					),
					"nan_canonicalization": hclspec.NewDefault(
						hclspec.NewAttr("nan_canonicalization", "bool", false),
						hclspec.NewLiteral(`false`),
					),
				})),
					hclspec.NewLiteral(`{
						optimize: 1,
						debug_verifier: false,
						nan_canonicalization: false,
					}`),
				),
				//"debug": hclspec.NewDefault(
				//	hclspec.NewAttr("default", "bool", false),
				//	hclspec.NewLiteral("false"),
				//),
				//"cache": hclspec.NewDefault(
				//	hclspec.NewAttr("cache", "bool", false),
				//	hclspec.NewLiteral("true"),
				//),
				//"simd": hclspec.NewDefault(
				//	hclspec.NewAttr("simd", "bool", false),
				//	hclspec.NewLiteral("false"),
				//),
				//"reference_types": hclspec.NewDefault(
				//	hclspec.NewAttr("reference_types", "bool", false),
				//	hclspec.NewLiteral("false"),
				//),
				//"multi_value": hclspec.NewDefault(
				//	hclspec.NewAttr("multi_value", "bool", false),
				//	hclspec.NewLiteral("false"),
				//),
				//"threads": hclspec.NewDefault(
				//	hclspec.NewAttr("threads", "bool", false),
				//	hclspec.NewLiteral("false"),
				//),
				//"bulk_memory": hclspec.NewDefault(
				//	hclspec.NewAttr("bulk_memory", "bool", false),
				//	hclspec.NewLiteral("false"),
				//),
				//"multi_memory": hclspec.NewDefault(
				//	hclspec.NewAttr("multi_memory", "bool", false),
				//	hclspec.NewLiteral("false"),
				//),
			})),
			hclspec.NewLiteral(`{
				strategy: "auto",
				cranelift_options: {
					optimize: 1,
					debug_verifier: false,
					nan_canonicalization: false,
				},
			}`),
		),
		"profiler": hclspec.NewDefault(
			hclspec.NewAttr("profiler", "string", false),
			hclspec.NewLiteral(`"none"`),
		),
		"dir_map": hclspec.NewAttr("port_map", "list(map(string))", false),
	})

	// capabilities indicates what optional features this driver supports
	// this should be set according to the target run time.
	capabilities = &drivers.Capabilities{
		// The plugin's capabilities signal Nomad which extra functionalities
		// are supported. For a list of available options check the docs page:
		// https://godoc.org/github.com/hashicorp/nomad/plugins/drivers#Capabilities
		SendSignals: true,
		Exec:        false,
		FSIsolation: drivers.FSIsolationNone,
		NetIsolationModes: []drivers.NetIsolationMode{
			drivers.NetIsolationModeHost,
		},
		MustInitiateNetwork: false,
		MountConfigs:        drivers.MountConfigSupportAll,
	}
)

// Config contains configuration information for the plugin
type Config struct {
	// This struct is the decoded version of the schema defined in the
	// configSpec variable above. It's used to convert the HCL configuration
	// passed by the Nomad agent into Go constructs.
}

type CraneLiftOptions struct {
	DebugVerifier       bool              `codec:"debug_verifier"`
	OptLevel            wasmtime.OptLevel `codec:"optimize"`
	NANCanonicalization bool              `codec:"nan_canonicalization"`
}

type WasmTimeCompiler struct {
	Strategy         string           `codec:"strategy"`
	CraneLiftOptions CraneLiftOptions `codec:"cranelift_options"`
}

// TaskConfig contains configuration information for a task that runs with
// this plugin
type TaskConfig struct {
	// This struct is the decoded version of the schema defined in the
	// taskConfigSpec variable above. It's used to convert the string
	// configuration for the task into Go constructs.
	Compiler WasmTimeCompiler `codec:"compiler"`
	Profiler string           `codec:"profiler"`
}
