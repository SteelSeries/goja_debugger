module github.com/mostafa/goja_debugger

go 1.16

require (
	github.com/dop251/goja v0.0.0-20231014103939-873a1496dc8e
	github.com/dop251/goja_nodejs v0.0.0-20231022114343-5c1f9037c9ab
	github.com/evanw/esbuild v0.12.13
	github.com/google/pprof v0.0.0-20231101202521-4ca4178f5c7a // indirect
	github.com/urfave/cli/v2 v2.3.0
)

replace github.com/dop251/goja => ../goja
