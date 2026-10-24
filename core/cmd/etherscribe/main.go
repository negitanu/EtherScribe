package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/negitanu/etherscribe/internal/export"
	"github.com/negitanu/etherscribe/internal/pipeline"
	"github.com/negitanu/etherscribe/internal/rpc"
	"github.com/negitanu/etherscribe/internal/version"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "version":
		info := version.Current()
		fmt.Printf("%s %s\n", info.Name, info.Version)
	case "rpc":
		if err := rpc.New(os.Stdin, os.Stdout).Run(); err != nil {
			fmt.Fprintf(os.Stderr, "rpc: %v\n", err)
			os.Exit(1)
		}
	case "convert":
		if err := runConvert(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "convert: %v\n", err)
			os.Exit(1)
		}
	case "compare":
		if err := runCompare(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "compare: %v\n", err)
			os.Exit(1)
		}
	case "sections":
		for _, k := range export.SectionKeys {
			fmt.Println(k)
		}
		fmt.Println("--- presets ---")
		for _, p := range export.Presets {
			fmt.Printf("%s\t%s\n", p.Key, p.Label)
		}
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `EtherScribe engine

Usage:
  etherscribe version
  etherscribe sections
  etherscribe convert -i <file> -o <out> [-f json|html|excel] [--preset NAME]
  etherscribe compare --before <file> --after <file> -o <out> [-f json|html]
  etherscribe rpc
`)
}

func runConvert(args []string) error {
	var in, out, format, preset string
	var sections []string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-i", "--input":
			i++
			in = args[i]
		case "-o", "--output":
			i++
			out = args[i]
		case "-f", "--format":
			i++
			format = args[i]
		case "--preset":
			i++
			preset = args[i]
		case "--sections":
			i++
			sections = splitCSV(args[i])
		default:
			return fmt.Errorf("unknown flag: %s", args[i])
		}
	}
	if in == "" || out == "" {
		return fmt.Errorf("-i and -o are required")
	}
	_, err := pipeline.ConvertFile(pipeline.ConvertRequest{
		Path: in, Output: out, Format: format, Preset: preset, Sections: sections,
	})
	return err
}

func runCompare(args []string) error {
	var before, after, out, format string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--before":
			i++
			before = args[i]
		case "--after":
			i++
			after = args[i]
		case "-o", "--output":
			i++
			out = args[i]
		case "-f", "--format":
			i++
			format = args[i]
		default:
			return fmt.Errorf("unknown flag: %s", args[i])
		}
	}
	if before == "" || after == "" || out == "" {
		return fmt.Errorf("--before, --after and -o are required")
	}
	_, err := pipeline.CompareFiles(pipeline.CompareRequest{
		Before: before, After: after, Output: out, Format: format,
	})
	return err
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
