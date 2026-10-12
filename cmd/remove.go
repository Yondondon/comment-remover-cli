/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"go/scanner"
	"go/token"
	"log"
	"os"

	"github.com/spf13/cobra"
)

type Comment struct {
	Start int
	End   int
}

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "remove short",
	Long:  `remove long`,
	Args:  cobra.ExactArgs(1),
	Run:   removeComments,
}

func init() {
	rootCmd.AddCommand(removeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// removeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// removeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func readFile(path string) []byte {
	src, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	return src
}

func writeFile(path string, src []byte) {
	err := os.WriteFile(path, src, 0666)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func scanComments(src []byte) []Comment {
	fset := token.NewFileSet()
	file := fset.AddFile("", fset.Base(), len(src))

	var s scanner.Scanner
	s.Init(file, src, nil, scanner.ScanComments)

	var comments []Comment

	for {
		pos, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}

		if tok == token.COMMENT {
			start := file.Offset(pos)
			comments = append(comments, Comment{
				Start: start,
				End:   start + len(lit),
			})
		}
	}

	return comments
}

func removeComments(cmd *cobra.Command, args []string) {
	path := args[0]
	src := readFile(path)
	comments := scanComments(src)

	var result bytes.Buffer
	cursor := 0

	for _, c := range comments {
		lineStart := bytes.LastIndexByte(src[:c.Start], '\n') + 1
		prefix := src[lineStart:c.Start]
		isWholeLine := len(bytes.TrimSpace(prefix)) == 0

		if isWholeLine {
			result.Write(src[cursor:lineStart])

			cursor = c.End
			for cursor < len(src) &&
				(src[cursor] == ' ' || src[cursor] == '\t' || src[cursor] == '\r') {
				cursor++
			}
			if cursor < len(src) && src[cursor] == '\n' {
				cursor++
			}
		} else {
			result.Write(src[cursor:c.Start])
			cursor = c.End
		}
	}

	result.Write(src[cursor:])
	writeFile(path, result.Bytes())
}
