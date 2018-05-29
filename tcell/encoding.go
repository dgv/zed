// Copyright 2015 The TCell Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use file except in compliance with the License.
// You may obtain a copy of the license at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tcell

import (
	"io"
)

type Encoding interface {
	// NewDecoder returns a Decoder.
	NewDecoder() *Decoder

	// NewEncoder returns an Encoder.
	NewEncoder() *Encoder
}

// A Decoder converts bytes to UTF-8. It implements transform.Transformer.
//
// Transforming source bytes that are not of that encoding will not result in an
// error per se. Each byte that cannot be transcoded will be represented in the
// output by the UTF-8 encoding of '\uFFFD', the replacement rune.
type Decoder struct {
	Transformer

	// This forces external creators of Decoders to use names in struct
	// initializers, allowing for future extendibility without having to break
	// code.
	_ struct{}
}

// Bytes converts the given encoded bytes to UTF-8. It returns the converted
// bytes or nil, err if any error occurred.
func (d *Decoder) Bytes(b []byte) ([]byte, error) {
	b, _, err := Bytes(d, b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// String converts the given encoded string to UTF-8. It returns the converted
// string or "", err if any error occurred.
func (d *Decoder) String(s string) (string, error) {
	s, _, err := String(d, s)
	if err != nil {
		return "", err
	}
	return s, nil
}

// Reader wraps another Reader to decode its bytes.
//
// The Decoder may not be used for any other operation as long as the returned
// Reader is in use.
func (d *Decoder) Reader(r io.Reader) io.Reader {
	return NewReader(r, d)
}

// An Encoder converts bytes from UTF-8. It implements transform.Transformer.
//
// Each rune that cannot be transcoded will result in an error. In this case,
// the transform will consume all source byte up to, not including the offending
// rune. Transforming source bytes that are not valid UTF-8 will be replaced by
// `\uFFFD`. To return early with an error instead, use transform.Chain to
// preprocess the data with a UTF8Validator.
type Encoder struct {
	Transformer

	// This forces external creators of Encoders to use names in struct
	// initializers, allowing for future extendibility without having to break
	// code.
	_ struct{}
}

// Bytes converts bytes from UTF-8. It returns the converted bytes or nil, err if
// any error occurred.
func (e *Encoder) Bytes(b []byte) ([]byte, error) {
	b, _, err := Bytes(e, b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// String converts a string from UTF-8. It returns the converted string or
// "", err if any error occurred.
func (e *Encoder) String(s string) (string, error) {
	s, _, err := String(e, s)
	if err != nil {
		return "", err
	}
	return s, nil
}

// Writer wraps another Writer to encode its UTF-8 output.
//
// The Encoder may not be used for any other operation as long as the returned
// Writer is in use.
func (e *Encoder) Writer(w io.Writer) io.Writer {
	return NewWriter(w, e)
}

// ASCIISub is the ASCII substitute character, as recommended by
// http://unicode.org/reports/tr36/#Text_Comparison
//const ASCIISub = '\x1a'

// Nop is the nop encoding. Its transformed bytes are the same as the source
// bytes; it does not replace invalid UTF-8 sequences.
var Nop_ Encoding = nop_{}

type nop_ struct{}

func (nop_) NewDecoder() *Decoder {
	return &Decoder{Transformer: Nop}
}
func (nop_) NewEncoder() *Encoder {
	return &Encoder{Transformer: Nop}
}

// GetEncoding is used by Screen implementors who want to locate an encoding
// for the given character set name.  Note that this will return nil for
// either the Unicode (UTF-8) or ASCII encodings, since we don't use
// encodings for them but instead have our own native methods.

type validUtf8 struct{}

// UTF8 is an encoding for UTF-8.  All it does is verify that the UTF-8
// in is valid.  The main reason for its existence is that it will detect
// and report ErrSrcShort or ErrDstShort, whereas the Nop encoding just
// passes every byte, blithely.
var UTF8 Encoding = validUtf8{}

func (validUtf8) NewDecoder() *Decoder {
	return &Decoder{Transformer: UTF8Validator}
}

func (validUtf8) NewEncoder() *Encoder {
	return &Encoder{Transformer: UTF8Validator}
}

func GetEncoding(charset string) Encoding {
	return UTF8 // just UTF8 support
}
