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

type XmlCodec struct{}

func (c XmlCodec) NewEncoder(w io.Writer) Encoder {
	return xml.NewEncoder(w)
}

func (c XmlCodec) NewDecoder(r io.Reader) Decoder {
	return xml.NewDecoder(r)
}

type JsonCodec struct{}

func (c JsonCodec) NewEncoder(w io.Writer) Encoder {
	return json.NewEncoder(w)
}

func (c JsonCodec) NewDecoder(r io.Reader) Decoder {
	return json.NewDecoder(r)
}

type GobCodec struct{}

func (c GobCodec) NewEncoder(w io.Writer) Encoder {
	return gob.NewEncoder(w)
}

func (c GobCodec) NewDecoder(r io.Reader) Decoder {
	return gob.NewDecoder(r)
}
