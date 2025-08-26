package main

import (
	"flag"
	"fmt"
	"os"
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

	generator := Generator{
		PackageName: *outputPkg,
		OutputDir:   *outputDir,
		Verbose:     *verbose,
	}

	var err error
	if *inputFile != "" {
		err = generator.ProcessFile(*inputFile)
	} else {
		err = generator.ProcessDirectory(*inputDir)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if *verbose {
		fmt.Println("Code generation completed successfully!")
	}
}
