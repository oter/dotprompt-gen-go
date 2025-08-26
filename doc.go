// Package main implements prompt_codegen, a code generation tool that creates
// Go request and response models from .prompt files in dotprompt format.
//
// The tool supports both Picoschema and JSON Schema formats and can be used
// with go:generate directives for automated code generation.
//
// Usage:
//
//	//go:generate prompt_codegen -dir prompts/ -pkg models -out models/
package main
