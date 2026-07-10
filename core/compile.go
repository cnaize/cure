package core

import (
	"bufio"
	"io"
	"iter"
	"slices"
	"strings"

	"github.com/sansecio/yargo/ast"
	"github.com/sansecio/yargo/parser"
	"github.com/sansecio/yargo/scanner"
)

var YaraRulePrefixes = []string{"rule ", "global rule ", "private rule "}

func Compile(files iter.Seq[io.Reader]) (*scanner.Rules, error) {
	var buff strings.Builder
	var rules ast.RuleSet
	parser := parser.New()
	for file := range files {
		buff.Reset()

		reader := bufio.NewScanner(file)
		for {
			stop := !reader.Scan()

			raw := reader.Text()
			line := strings.TrimSpace(raw)
			if stop || slices.ContainsFunc(YaraRulePrefixes, func(prefix string) bool {
				return strings.HasPrefix(line, prefix)
			}) {
				if buff.Len() > 0 {
					if rule, err := parser.Parse(buff.String()); err == nil {
						rules.Rules = append(rules.Rules, rule.Rules...)
					}
					buff.Reset()
				}
			}

			if stop {
				break
			}

			buff.WriteString(raw + "\n")
		}
	}

	return scanner.CompileWithOptions(&rules, scanner.CompileOptions{SkipInvalidRegex: true})
}
