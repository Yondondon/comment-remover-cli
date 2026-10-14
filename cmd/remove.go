package cmd

import (
	"bytes"
	"fmt"
	"go/format"
	"go/scanner"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

type commentSpan struct{ start, end int }

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove comments from .go files",
	Long:  "Remove comments from a .go file, or from all .go files in a directory.",
	Args:  cobra.NoArgs,
	RunE:  runRemove,
}

func init() {
	rootCmd.AddCommand(removeCmd)

	f := removeCmd.Flags()
	f.BoolP("recursive", "r", false, "recursively checks subdirectories")
	f.StringP("path", "p", "", "(required) path to file or directory")
	_ = removeCmd.MarkFlagRequired("path")
}

func runRemove(cmd *cobra.Command, _ []string) error {
	path, err := cmd.Flags().GetString("path")
	if err != nil {
		return err
	}
	recursive, err := cmd.Flags().GetBool("recursive")
	if err != nil {
		return err
	}

	files, err := collectGoFiles(path, recursive)
	if err != nil {
		return err
	}

	var g errgroup.Group
	g.SetLimit(runtime.NumCPU())
	for _, file := range files {
		g.Go(func() error { return processFile(file) })
	}
	if err := g.Wait(); err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Processed %d file(s).\n", len(files))
	return nil
}

func collectGoFiles(root string, recursive bool) ([]string, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		if filepath.Ext(root) != ".go" {
			return nil, fmt.Errorf("%s is not a .go file", root)
		}
		return []string{root}, nil
	}

	var files []string
	err = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if p != root && !recursive {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(p) == ".go" {
			files = append(files, p)
		}
		return nil
	})
	return files, err
}

func isDirective(lit string) bool {
	return strings.HasPrefix(lit, "//go:") || strings.HasPrefix(lit, "// +build")
}

func scanComments(src []byte) []commentSpan {
	fset := token.NewFileSet()
	file := fset.AddFile("", fset.Base(), len(src))

	var s scanner.Scanner
	s.Init(file, src, nil, scanner.ScanComments)

	var spans []commentSpan
	for {
		pos, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}
		if tok == token.COMMENT && !isDirective(lit) {
			start := file.Offset(pos)
			spans = append(spans, commentSpan{start, start + len(lit)})
		}
	}
	return spans
}

func stripComments(src []byte) ([]byte, error) {
	spans := scanComments(src)
	if len(spans) == 0 {
		return src, nil
	}

	var out bytes.Buffer
	cursor := 0

	for _, c := range spans {
		lineStart := bytes.LastIndexByte(src[:c.start], '\n') + 1
		wholeLine := len(bytes.TrimSpace(src[lineStart:c.start])) == 0

		if !wholeLine {
			out.Write(src[cursor:c.start])
			cursor = c.end
			continue
		}

		out.Write(src[cursor:lineStart])
		cursor = c.end
		for cursor < len(src) && (src[cursor] == ' ' || src[cursor] == '\t' || src[cursor] == '\r') {
			cursor++
		}
		if cursor < len(src) && src[cursor] == '\n' {
			cursor++
		}
	}
	out.Write(src[cursor:])

	return format.Source(out.Bytes())
}

func processFile(path string) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	result, err := stripComments(src)
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	if bytes.Equal(src, result) {
		return nil
	}

	return os.WriteFile(path, result, 0o644)
}
