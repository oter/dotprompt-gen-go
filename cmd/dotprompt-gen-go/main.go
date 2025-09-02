// Package main implements dotprompt-gen, a code generation tool that creates
// Go request and response models from .prompt files in dotprompt format.
//
// The tool supports both Picoschema and JSON Schema formats and can be used
// to generate type-safe Go structs from prompt definitions.
//
// Usage:
//
//	dotprompt-gen -file path/to/prompt.prompt
//	dotprompt-gen -dir path/to/prompts/ -pkg models -out ./generated/
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/oter/dotprompt-gen-go/internal/codegen"
	"github.com/oter/dotprompt-gen-go/internal/generator"
)

func main() {
	var (
		inputFile = flag.String("file", "", "Single .prompt file to process")
		inputDir  = flag.String("dir", "", "Directory containing .prompt files")
		outputPkg = flag.String("pkg", "models", "Output package name")
		outputDir = flag.String("out", "", "Output directory (default: same as input)")
		verbose   = flag.Bool("v", false, "Verbose output")
		help      = flag.Bool("h", false, "Show help")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Generate Go request/response models from dotprompt files.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(
			os.Stderr,
			"  %s -file app/classify/prompts/classify_habits.prompt\n",
			os.Args[0],
		)
		fmt.Fprintf(os.Stderr, "  %s -dir app/classify/prompts/ -pkg models\n", os.Args[0])
		fmt.Fprintf(
			os.Stderr,
			"  %s -dir app/classify/prompts/ -out app/classify/models/\n",
			os.Args[0],
		)
	}

	flag.Parse()

	if *help {
		flag.Usage()

		return
	}

	if *inputFile == "" && *inputDir == "" {
		fmt.Fprintf(os.Stderr, "Error: Either -file or -dir must be specified\n\n")
		flag.Usage()
		os.Exit(1)
	}

	if *inputFile != "" && *inputDir != "" {
		fmt.Fprintf(os.Stderr, "Error: Cannot specify both -file and -dir\n\n")
		flag.Usage()
		os.Exit(1)
	}

	gen := codegen.Generator{
		PackageName: *outputPkg,
		OutputDir:   *outputDir,
		Verbose:     *verbose,
	}

	var err error
	if *inputFile != "" {
		err = generator.ProcessFile(gen, *inputFile)
	} else {
		err = generator.ProcessDirectory(gen, *inputDir)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if *verbose {
		fmt.Println("Code generation completed successfully!")
	}
}
