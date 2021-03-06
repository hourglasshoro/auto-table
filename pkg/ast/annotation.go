package ast

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/hourglasshoro/auto-table/pkg/utils"
	"go/ast"
	"strconv"
	"strings"
)

type annotation struct {
	Table  string
	Option string
}

const (
	commentPrefix       = "//"
	annotationSeparator = ':'
)

func parseAnnotation(g *ast.CommentGroup, marker string) (*annotation, error) {
	for _, c := range g.List {
		if !strings.HasPrefix(c.Text, commentPrefix) {
			continue
		}
		s := strings.TrimSpace(c.Text[len(commentPrefix):])
		if !strings.HasPrefix(s, marker) {
			continue
		}
		if len(s) == len(marker) {
			return &annotation{}, nil
		}
		if !utils.IsSpace(s[len(marker)]) {
			continue
		}
		var a annotation
		scanner := bufio.NewScanner(strings.NewReader(s[len(marker):]))
		scanner.Split(splitAnnotationTags)
		for scanner.Scan() {
			ss := strings.SplitN(scanner.Text(), string(annotationSeparator), 2)
			switch k, v := ss[0], ss[1]; k {
			case "table":
				s, err := parseString(v)
				if err != nil {
					return nil, fmt.Errorf("auto-table: BUG: %v", err)
				}
				a.Table = s
			case "option":
				s, err := parseString(v)
				if err != nil {
					return nil, fmt.Errorf("auto-table: BUG: %v", err)
				}
				a.Option = s
			default:
				return nil, fmt.Errorf("auto-table: unsupported annotation: %v", k)
			}
		}
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("%v: %v", err, c.Text)
		}
		return &a, nil
	}
	return nil, nil
}

func splitAnnotationTags(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF {
		return 0, nil, nil
	}
	for ; advance < len(data); advance++ {
		if !utils.IsSpace(data[advance]) {
			break
		}
	}
	i := bytes.IndexByte(data[advance:], annotationSeparator)
	if i < 1 {
		return 0, nil, fmt.Errorf("auto-table: invalid annotation")
	}
	advance += i + 1
	if advance >= len(data) {
		return 0, nil, fmt.Errorf("auto-table: invalid annotation")
	}
	switch quote := data[advance]; quote {
	case '"':
		for advance++; advance < len(data); advance++ {
			i := bytes.IndexByte(data[advance:], quote)
			if i < 0 {
				break
			}
			advance += i
			if data[advance-1] != '\\' {
				return advance + 1, bytes.TrimSpace(data[:advance+1]), nil
			}
		}
		return 0, nil, fmt.Errorf("auto-table: invalid annotation: string not terminated")
	case '`':
		for advance++; advance < len(data); {
			i := bytes.IndexByte(data[advance:], quote)
			if i < 0 {
				break
			}
			advance += i
			return advance + 1, bytes.TrimSpace(data[:advance+1]), nil
		}
		return 0, nil, fmt.Errorf("auto-table: invalid annotation: string not terminated")
	}
	if utils.IsSpace(data[advance]) {
		return 0, nil, fmt.Errorf("auto-table: invalid annotation: value not given")
	}
	for advance++; advance < len(data); advance++ {
		if utils.IsSpace(data[advance]) {
			return advance, bytes.TrimSpace(data[:advance]), nil
		}
	}
	return advance, bytes.TrimSpace(data[:advance]), nil
}

func parseString(s string) (string, error) {
	if b := s[0]; b == '"' || b == '`' {
		return strconv.Unquote(s)
	}
	return s, nil
}
