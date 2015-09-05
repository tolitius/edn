// Copyright 2015 Jean Niklas L'orange.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package edn

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

type Keyword string

func (k Keyword) String() string {
	return fmt.Sprintf(":%s", string(k))
}

func (k Keyword) MarshalEDN() ([]byte, error) {
	return []byte(k.String()), nil
}

type Symbol string

func (s Symbol) String() string {
	return string(s)
}

func (s Symbol) MarshalEDN() ([]byte, error) {
	return []byte(s), nil
}

type Vector []interface{}

func (v Vector) String() string {
	vals := []string{}
	for _, val := range v {
		vals = append(vals, fmt.Sprint(val))
	}
	return "[" + strings.Join(vals, " ") + "]"
}

type List []interface{}

func (l List) String() string {
	vals := []string{}
	for _, val := range l {
		vals = append(vals, fmt.Sprint(val))
	}
	return "(" + strings.Join(vals, " ") + ")"
}

type Tag struct {
	Tagname string
	Value   interface{}
}

func (t Tag) String() string {
	return fmt.Sprintf("#%s %s", t.Tagname, t.Value)
}

func (t Tag) MarshalEDN() ([]byte, error) {
	str := []byte(fmt.Sprintf(`#%s `, t.Tagname))
	b, err := Marshal(t.Value)
	if err != nil {
		return nil, err
	}
	return append(str, b...), nil
}

func (t *Tag) UnmarshalEDN(bs []byte) error {
	// read actual tag, using the lexer.
	var lex lexer
	lex.reset()
	buf := bufio.NewReader(bytes.NewBuffer(bs))
	start := 0
	endTag := 0
tag:
	for {
		r, rlen, err := buf.ReadRune()
		if err != nil {
			return err
		}

		ls := lex.state(r)
		switch ls {
		case lexIgnore:
			start += rlen
			endTag += rlen
		case lexError:
			return lex.err
		case lexEndPrev:
			break tag
		case lexEnd: // unexpected, assuming tag which is not ending with lexEnd
			return errUnexpected
		case lexCont:
			endTag += rlen
		}
	}
	t.Tagname = string(bs[start+1 : endTag])
	return Unmarshal(bs[endTag:], &t.Value)
}

type Rune rune

func (r Rune) MarshalEDN() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, 10))
	encodeRune(buf, rune(r))
	return buf.Bytes(), nil
}

func encodeRune(buf *bytes.Buffer, r rune) {
	if !isWhitespace(r) {
		buf.WriteByte('\\')
		buf.WriteRune(r)
	} else {
		switch r {
		case '\b':
			buf.WriteString(`\backspace`)
		case '\f':
			buf.WriteString(`\formfeed`)
		case '\n':
			buf.WriteString(`\newline`)
		case '\r':
			buf.WriteString(`\return`)
		case '\t':
			buf.WriteString(`\tab`)
		case ' ':
			buf.WriteString(`\space`)
		default:
			val := strconv.QuoteRuneToASCII(r)
			buf.WriteString(val[1 : len(val)-1])
		}
	}
}

type Set map[interface{}]bool

func (s Set) String() string {
	keys := []string{}
	for k, _ := range s {
		keys = append(keys, fmt.Sprint(k))
	}
	return "#{" + strings.Join(keys, " ") + "}"
}

type Map map[interface{}]interface{}

func (m Map) String() string {
	kvs := []string{}
	for k, v := range m {
		kvs = append(kvs, fmt.Sprintf("%s %s", k, v))
	}
	return "{" + strings.Join(kvs, ", ") + "}"
}
