package adapter

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"slices"
	"strings"
)

var crsFileExtensions = []string{".conf"}

// WARNING: not fully implemented yet
// developed to support OWASP CRS (https://github.com/coreruleset/coreruleset)
var _ Adapter = (*Crs)(nil)

type Crs struct {
	num int
}

func NewCrs() *Crs {
	return &Crs{}
}

func (a *Crs) Check(name string) bool {
	return slices.Contains(crsFileExtensions, strings.ToLower(filepath.Ext(name)))
}

func (a *Crs) Adapt(name string, in io.Reader) io.Reader {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".data":
		return a.adaptData(in)
	case ".conf":
		return a.adaptConf(in)
	}

	return in
}

func (a *Crs) adaptData(in io.Reader) io.Reader {
	var num int
	var buff strings.Builder
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) < 1 || strings.HasPrefix(line, "#") {
			continue
		}

		line = strings.ReplaceAll(line, "\\", "\\\\")
		line = strings.ReplaceAll(line, "\"", "\\\"")

		fmt.Fprintf(&buff, "\t\t\t$s%d = \"%s\" nocase ascii\n", num, line)
		num++
	}

	out := fmt.Sprintf(`
	rule CRS_data_%d {
		strings:
%s
		condition:
			any of them
	}`+"\n", a.num, buff.String())
	a.num++

	return strings.NewReader(out)
}

func (a *Crs) adaptConf(in io.Reader) io.Reader {
	var (
		rxTag      = []byte("@rx")
		noCaseTag  = []byte("(?i)")
		chainTag   = []byte("chain")
		includeTag = []byte("tag:'attack-")
		excludeTag = []byte("tag:'attack-protocol'")
	)

	data, err := io.ReadAll(in)
	if err != nil {
		return nil
	}

	var out bytes.Buffer
	var rule []byte
	tail := data
	for {
		rule, tail = getCrsRule(tail)
		if len(rule) < 1 {
			break
		}

		for bytes.Contains(rule, chainTag) {
			rule, tail = getCrsRule(tail)
			if !bytes.Contains(rule, chainTag) {
				rule, tail = getCrsRule(tail)
			}
		}

		if !bytes.Contains(rule, includeTag) || bytes.Contains(rule, excludeTag) {
			continue
		}

		head := bytes.SplitN(rule, []byte{'\n'}, 2)[0]
		head = bytes.TrimSuffix(head, []byte{'\\'})
		head = bytes.TrimSuffix(bytes.TrimSpace(head), []byte{'"'})

		pre, rx, ok := bytes.Cut(head, rxTag)
		if !ok {
			continue
		}

		pre = bytes.TrimSpace(pre)
		if pre[len(pre)-1] == '!' {
			continue
		}

		rx = bytes.TrimSpace(rx)
		if len(rx) < 3 {
			continue
		}

		modifier := "ascii"
		if bytes.HasPrefix(rx, noCaseTag) {
			rx = bytes.Replace(rx, noCaseTag, nil, 1)
			modifier = "nocase ascii"
		}

		var buff bytes.Buffer
		for i, c := range rx {
			if c == '/' {
				if i == 0 || rx[i-1] != '\\' {
					buff.WriteByte('\\')
				}
			}
			buff.WriteByte(c)
		}

		fmt.Fprintf(&out, `
		rule CRS_conf_%d {
			strings:
				$s = /%s/ %s
			condition:
				$s
		}`+"\n", a.num, buff.Bytes(), modifier)
		a.num++
	}

	return bytes.NewReader(out.Bytes())
}

func getCrsRule(data []byte) ([]byte, []byte) {
	var secRuleTag = []byte("SecRule ")

	isr := bytes.Index(data, secRuleTag)
	if isr < 0 {
		return nil, nil
	}

	data = data[isr:]
	var head, tail []byte

	isr = bytes.Index(data[len(secRuleTag):], secRuleTag)
	if isr < 0 {
		head, tail = data, nil
	} else {
		head, tail = data[:len(secRuleTag)+isr], data[len(secRuleTag)+isr:]
	}

	var rule []byte
	for line := range bytes.SplitSeq(head, []byte{'\n'}) {
		line = bytes.TrimSpace(line)
		if len(line) < 1 || bytes.HasPrefix(line, []byte{'#'}) {
			continue
		}

		rule = append(rule, line...)
		rule = append(rule, '\n')
	}

	return rule, tail
}
