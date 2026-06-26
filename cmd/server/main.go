// Package main is the entrypoint for the RedIntel Sentinel platform.
//
// RedIntel Sentinel is an enterprise Attack Surface Management (ASM) platform
// intended strictly for authorized, defensive security assessments. The binary
// is a thin wrapper around the Cobra-powered CLI defined in internal/cli.
package main

import "github.com/Skypieee6/redintel-sentinel/internal/cli"

func main() {
	cli.Execute()
}
