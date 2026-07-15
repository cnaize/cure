package adapter

import (
	"bufio"
	"fmt"
	"io"
	"path/filepath"
	"slices"
	"strings"
)

var CrsFileExtensions = []string{".conf"}

type Crs struct {
}

func NewCrs() Crs {
	return Crs{}
}

func (a Crs) Check(name string) bool {
	return slices.Contains(CrsFileExtensions, strings.ToLower(filepath.Ext(name)))
}

func (a Crs) Adapt(in io.Reader) io.Reader {
	var out strings.Builder
	var buff strings.Builder
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) < 1 || strings.HasPrefix(line, "#") {
			continue
		}

		if before, ok := strings.CutSuffix(line, "\\"); ok {
			buff.WriteString(before)
			buff.WriteString(" ")
			continue
		}

		buff.WriteString(line)
		rule := buff.String()
		buff.Reset()

		_, rx, ok := strings.Cut(rule, "@rx")
		if !ok {
			continue
		}

		rx = strings.TrimSpace(rx)
		if len(rx) > 1 {
			if (rx[0] == '"' && rx[len(rx)-1] == '"') || (rx[0] == '\'' && rx[len(rx)-1] == '\'') {
				rx = rx[1 : len(rx)-1]
			}
		}

		var nocase bool
		if strings.HasPrefix(rx, "(?i)") {
			rx = strings.Replace(rx, "(?i)", "", 1)
			nocase = true
		}

		for i, s := range rx {
			if s == '/' {
				if i == 0 || rx[i-1] != '\\' {
					buff.WriteByte('\\')
				}
			}
			buff.WriteRune(s)
		}

		modifier := "ascii"
		if nocase {
			modifier = "nocase ascii"
		}

		fmt.Fprintf(&out, `
		rule CRS {
			strings:
				$s = /%s/ %s
			condition:
				$s
		}`, buff.String(), modifier)
		buff.Reset()
	}

	return strings.NewReader(out.String())
}
