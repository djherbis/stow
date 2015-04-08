package stow

import (
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"io"
)

type Codec interface {
	NewEncoder(io.Writer) Encoder
	NewDecoder(io.Reader) Decoder
}

type Encoder interface {
	Encode(interface{}) error
}

type Decoder interface {
	Decode(interface{}) error
}

var (
	_ Codec = XMLCodec{}
	_ Codec = JSONCodec{}
	_ Codec = GobCodec{}
)

type XMLCodec struct{}

func (c XMLCodec) NewEncoder(w io.Writer) Encoder {
	return xml.NewEncoder(w)
}

func (c XMLCodec) NewDecoder(r io.Reader) Decoder {
	return xml.NewDecoder(r)
}

type JSONCodec struct{}

func (c JSONCodec) NewEncoder(w io.Writer) Encoder {
	return json.NewEncoder(w)
}

func (c JSONCodec) NewDecoder(r io.Reader) Decoder {
	return json.NewDecoder(r)
}

type GobCodec struct{}

func (c GobCodec) NewEncoder(w io.Writer) Encoder {
	return gob.NewEncoder(w)
}

func (c GobCodec) NewDecoder(r io.Reader) Decoder {
	return gob.NewDecoder(r)
}
