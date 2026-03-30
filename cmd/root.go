package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fantods/gjq/internal/output"
	"github.com/fantods/gjq/internal/query"
	"github.com/spf13/cobra"
)

var (
	flagIgnoreCase  bool
	flagCompact     bool
	flagCount       bool
	flagDepth       bool
	flagNoDisplay   bool
	flagFixedString bool
	flagWithPath    bool
	flagNoPath      bool
)

var rootCmd = &cobra.Command{
	Use:   "gjq [OPTIONS] [QUERY] [FILE]",
	Short: "A JSONPath-inspired query language for JSON documents",
	Long: `gjq is a CLI tool for querying JSON documents using regular path queries.
Queries are regular expressions applied to JSON paths — matching keys and array
indices rather than characters.`,
	Version:       Version,
	SilenceUsage:  true,
	SilenceErrors: true,
	Args:          cobra.MaximumNArgs(2),
	RunE:          runRoot,
}

func init() {
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)

	rootCmd.Flags().BoolVarP(&flagIgnoreCase, "ignore-case", "i", false, "Case insensitive search")
	rootCmd.Flags().BoolVar(&flagCompact, "compact", false, "Do not pretty-print the JSON output")
	rootCmd.Flags().BoolVar(&flagCount, "count", false, "Display count of number of matches")
	rootCmd.Flags().BoolVar(&flagDepth, "depth", false, "Display depth of the input document")
	rootCmd.Flags().BoolVarP(&flagNoDisplay, "no-display", "n", false, "Do not display matched JSON values")
	rootCmd.Flags().BoolVarP(&flagFixedString, "fixed-string", "F", false, "Treat the query as a literal field name and search at any depth")
	rootCmd.Flags().BoolVar(&flagWithPath, "with-path", false, "Always print the path header, even when output is piped")
	rootCmd.Flags().BoolVar(&flagNoPath, "no-path", false, "Never print the path header, even in a terminal")

	rootCmd.MarkFlagsMutuallyExclusive("with-path", "no-path")
}

func Execute() error {
	return rootCmd.Execute()
}

func runRoot(cmd *cobra.Command, args []string) error {
	var queryStr string
	var fileArg string

	switch len(args) {
	case 2:
		queryStr = args[0]
		fileArg = args[1]
	case 1:
		queryStr = args[0]
	case 0:
		fmt.Fprintln(cmd.ErrOrStderr(), "Error: query string required unless using subcommand")
		return fmt.Errorf("query string required")
	}

	var q query.Query
	if flagFixedString {
		q = query.NewSequence([]query.Query{
			query.NewKleeneStar(query.NewDisjunction([]query.Query{
				query.NewFieldWildcard(),
				query.NewArrayWildcard(),
			})),
			query.NewField(queryStr),
		})
	} else {
		var err error
		q, err = query.ParseQuery(queryStr)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err.Error())
			return err
		}
	}

	data, err := readInput(fileArg, cmd)
	if err != nil {
		return fmt.Errorf("reading input: %w", err)
	}

	root, err := query.ParseJSON(string(data))
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err.Error())
		return err
	}

	dfa := query.NewQueryDFA(&q, flagIgnoreCase)
	results := dfa.Find(root)

	showPath := resolveShowPath()
	colorize := isTerminal(os.Stdout)

	w := bufio.NewWriterSize(cmd.OutOrStdout(), 4096)

	if flagCount {
		fmt.Fprintf(w, "Found matches: %d\n", len(results))
	}

	if flagDepth {
		fmt.Fprintf(w, "Depth: %d\n", output.Depth(root))
	}

	if !flagNoDisplay {
		pretty := !flagCompact
		for _, result := range results {
			output.WriteResult(w, result.Value, result.Path, pretty, showPath, colorize)
		}
	}

	if err := w.Flush(); err != nil {
		if strings.Contains(err.Error(), "broken pipe") {
			return nil
		}
		return err
	}

	return nil
}

func readInput(fileArg string, cmd *cobra.Command) ([]byte, error) {
	if fileArg != "" {
		data, err := os.ReadFile(fileArg)
		if err != nil {
			return nil, fmt.Errorf("failed to open file %s: %w", fileArg, err)
		}
		return data, nil
	}

	if isTerminal(os.Stdin) {
		fmt.Fprintln(cmd.ErrOrStderr(), "Error: no input specified (provide a file or pipe JSON to stdin)")
		return nil, fmt.Errorf("no input specified")
	}

	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("reading stdin: %w", err)
	}
	return data, nil
}

func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

func resolveShowPath() bool {
	if flagWithPath {
		return true
	}
	if flagNoPath {
		return false
	}
	return isTerminal(os.Stdout)
}
