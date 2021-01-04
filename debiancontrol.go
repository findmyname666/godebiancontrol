// vim:ts=4:sw=4:noexpandtab
// © 2012-2014 Michael Stapelberg (see also: LICENSE)

// Package debiancontrol implements a parser for Debian control files.
package godebiancontrol

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"unicode"
)

type FieldType int

const (
	Simple FieldType = iota
	Folded
	Multiline
)

var fieldType = make(map[string]FieldType)

type Paragraph struct {
	Fields map[string]string
	Order []string
}

func init() {
	fieldType["Description"] = Multiline
	fieldType["Files"] = Multiline
	fieldType["Changes"] = Multiline
	fieldType["Checksums-Sha1"] = Multiline
	fieldType["Checksums-Sha256"] = Multiline
	fieldType["Package-List"] = Multiline
	fieldType["MD5Sum"] = Multiline
	fieldType["SHA1"] = Multiline
	fieldType["SHA256"] = Multiline
}

// Parses a Debian control file and returns a slice of Paragraphs.
//
// Implemented according to chapter 5.1 (Syntax of control files) of the Debian
// Policy Manual:
// http://www.debian.org/doc/debian-policy/ch-controlfields.html
func Parse(input io.Reader) (paragraphs []Paragraph, err error) {
	reader := bufio.NewReader(input)
	var paragraph Paragraph

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		trimmed := strings.TrimSpace(line)

		// Check if the line is empty (or consists only of whitespace). This
		// marks a new paragraph.
		if trimmed == "" {
			if len(paragraph.Fields) > 0 {
				paragraphs = append(paragraphs, paragraph)
			}
			paragraph = Paragraph{}
			continue
		}

		// folded and multiline fields must start with a space or tab.
		if line[0] == ' ' || line[0] == '\t' {
			lastKey := paragraph.Order[len(paragraph.Order) - 1]
			if fieldType[lastKey] == Multiline {
				// Whitespace, including newlines, is significant in the values
				// of multiline fields, therefore we just append the line
				// as-is.
				paragraph.Fields[lastKey] += line
			} else {
				// For folded lines we strip whitespace before appending.
				paragraph.Fields[lastKey] += trimmed
			}
		} else {
			split := strings.Split(trimmed, ":")
			key := split[0]
			value := strings.TrimLeftFunc(trimmed[len(key)+1:], unicode.IsSpace)
			paragraph.Fields[key] = value
			paragraph.Order = append(paragraph.Order, key)
		}
	}

	// Append last paragraph, but only if it is non-empty.
	// The case of an empty last paragraph happens when the file ends with a
	// blank line.
	if len(paragraph.Fields) > 0 {
		paragraphs = append(paragraphs, paragraph)
	}

	return
}

// Convert paragraph to string.
func (p *Paragraph) String() string {
	var buff bytes.Buffer

	for _, k := range p.Order {
		v, ok := p.Fields[k]

		// skip keys without values
		if !ok {
			continue
		}

		// Multiline fields contains line endings.
		// To make it uniform let's strip them and add one.
		vTrimmed := strings.TrimRight(v, "\n")
		buff.WriteString(k + ": " + vTrimmed + "\n")
	}
	return buff.String()
}

// Insert k, v fields into paragraph map. If k already exists v is replaced.
func (p *Paragraph) Set(k, v string) {
	p.Fields[k] = v
	p.Order = append(p.Order, k)
}

func (p *Paragraph) Get(k string) string {
	return p.Fields[k]
}

func (p *Paragraph) Len() int {
	return len(p.Fields)
}

func (p *Paragraph) Del(k string) {
	if _, ok := p.Fields[k]; !ok {
		return
	}

	delete(p.Fields, k)
	delItemSlice(k, p.Order)
}

func delItemSlice(x string, s []string) []string {
	i := getItemPositionSlice(x, s)
	return delIndexSlice(i, s)
}

func getItemPositionSlice(x string, s []string) int {
	for i, v := range s {
		if x == v {
			return i
		}
	}
	return -1
}

func delIndexSlice(i int, s []string) []string {
	return append(s[:i], s[i+1:]...)
}
