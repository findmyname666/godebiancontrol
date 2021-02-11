// vim:ts=4:sw=4:noexpandtab
// © 2012-2014 Michael Stapelberg (see also: LICENSE)

// Package debiancontrol implements a parser for Debian files
// such package control file, mirror Packages and Release files.
package godebiancontrol

import (
	"bufio"
	"bytes"
	"fmt"
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

type Paragraph struct {
	Fields    map[string]string
	FieldType map[string]FieldType
	Order     []string
}

// Initialize empty paragraph.
func NewParagraph() (p Paragraph) {
	p.Fields = make(map[string]string)
	p.FieldType = make(map[string]FieldType)
	p.Order = make([]string, 0)

	return
}

// Parses a Debian control file and returns a slice of Paragraphs.
//
// Implemented according to chapter 5.1 (Syntax of control files) of the Debian
// Policy Manual:
// http://www.debian.org/doc/debian-policy/ch-controlfields.html
func Parse(input io.Reader) (paragraphs []Paragraph, err error) {
	reader := bufio.NewReader(input)
	paragraph := NewParagraph()

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
		// indicate a new paragraph.
		if trimmed == "" {
			if paragraph.Len() > 0 {
				paragraphs = append(paragraphs, paragraph)
			}
			paragraph = NewParagraph()
			continue
		}

		// Folded and multiline fields must start with a space or tab.
		if line[0] == ' ' || line[0] == '\t' {
			// Let's keep the lines in the original state.
			// Last key can be empty when paragraph is empty but it doesn't
			// doesn't affect our use-case.
			lastKey := paragraph.LastKey()
			lPrev := paragraph.Get(lastKey)
			paragraph.Set(lastKey, lPrev + line)
			continue
		}

		k, v := extractKV(line)
		// Empty value for a field indicate multiline value.
		if v == "" {
			paragraph.FieldType[k] = Multiline
		}

		paragraph.Set(k, v)
	}

	// Append last paragraph, but only if it is non-empty.
	// The last paragraph can be empty when the file ends with a
	// blank line.
	if paragraph.Len() > 0 {
		paragraphs = append(paragraphs, paragraph)
	}

	// Trim values from right.
	for _, p := range paragraphs {
		for k, v := range p.Fields {
			p.Set(k, strings.TrimRightFunc(v, unicode.IsSpace))
		}
	}

	return
}

// Extract field and its value from line.
func extractKV(l string) (k, v string) {
	split := strings.Split(l, ":")
	k = split[0]
	v = strings.TrimLeftFunc(l[len(k)+1:], unicode.IsSpace)

	return
}

// Convert paragraph to bytes buffer.
func (p *Paragraph) Bytes() *bytes.Buffer {
	var buff bytes.Buffer

	for _, k := range p.Order {
		v := p.Get(k)

		// skip keys without values
		if v == "" {
			continue
		}

		// Multiline fields contains line endings.
		// To make it uniform let's strip them and add one.
		vTrimmed := strings.TrimRight(v, "\n")

		buff.WriteString(k + ":")
		if p.FieldType[k] == Multiline {
			buff.WriteString("\n")
		} else {
			buff.WriteString(" ")
		}
		buff.WriteString(vTrimmed + "\n")
	}
	return &buff
}

// Convert paragraph to string.
func (p *Paragraph) String() string {
	buff := p.Bytes()

	return buff.String()
}

// Convert paragraphs to bytes buffer. Each paragraph is separated with
// the new line.
func ParagraphsToBytes(paragraphs []Paragraph) *bytes.Buffer {
	var buff bytes.Buffer
	paragraphsCount := len(paragraphs) - 1

	for i, p := range paragraphs {
		fmt.Fprint(&buff, p.String())

		// Don't append the new line after the last paragraph.
		if i == paragraphsCount {
			continue
		}

		buff.WriteString("\n")
	}

	return &buff
}

// Convert paragraphs to multi-line string. Each paragraph is separated with
// the new line.
func ParagraphsToText(paragraphs []Paragraph) string {
	buff := ParagraphsToBytes(paragraphs)

	return buff.String()
}

// Insert field (k) and its value (v) into paragraph map.
// If field exists already its value is replaced.
func (p *Paragraph) Set(k, v string) {
	vOld := p.Get(k)

	if vOld == "" {
		p.Order = append(p.Order, k)
	}

	p.Set(k, v)
}

// Get field value based on field name (k).
func (p *Paragraph) Get(k string) string {
	return p.Fields[k]
}

// Get number of fields in a paragraph.
func (p *Paragraph) Len() int {
	return len(p.Fields)
}

// Delete field from a paragraph based on the field name (k).
func (p *Paragraph) Del(k string) {
	if _, ok := p.Fields[k]; !ok {
		return
	}

	delete(p.Fields, k)
	delItemSlice(k, p.Order)
}

// Retrieve last key without it's removal.
func (p *Paragraph) LastKey() string {
	if len(p.Order) == 0 {
		return ""
	}
	return p.Order[len(p.Order)-1]
}

// Delete item (x) from slice (s).
func delItemSlice(x string, s []string) []string {
	i := getItemPositionSlice(x, s)
	return delIndexSlice(i, s)
}

// Get item (x) index position in a slice (s).
func getItemPositionSlice(x string, s []string) int {
	for i, v := range s {
		if x == v {
			return i
		}
	}
	return -1
}

// Delete item from slice (s) based on item index position.
func delIndexSlice(i int, s []string) []string {
	return append(s[:i], s[i+1:]...)
}
