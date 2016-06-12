package stow

import (
	"bytes"
	"io"
	"io/ioutil"
)

type delegateEncoder struct {
	Encoder
	io.Writer
}

type delegateDecoder struct {
	Decoder
	io.Reader
}

type primedCodec struct {
	codec Codec
	types []interface{}
	data  []byte
}

// NewPrimedCodec delegates to the passed codec for creating Encoders/Decoders.
// Newly created Encoder/Decoders will Encode/Decode the passed sample structs without
// actually writing/reading from their respective Writer/Readers. This is useful for
// Codec's like GobCodec{} which encodes/decodes extra type information whenever it
// sees a new type. Pass sample values for types you plan on Encoding/Decoding to this
// method in order to avoid the storage overhead of encoding their type informaton for every
// NewEncoder/NewDecoder.
func NewPrimedCodec(codec Codec, types ...interface{}) (Codec, error) {
	var buf bytes.Buffer
	enc := codec.NewEncoder(&buf)
	if err := enc.Encode(types); err != nil {
		return nil, err
	}
	var testTypes []interface{}
	if err := codec.NewDecoder(bytes.NewReader(buf.Bytes())).Decode(&testTypes); err != nil {
		return nil, err
	}
	return &primedCodec{
		codec: codec,
		types: types,
		data:  buf.Bytes(),
	}, nil
}

func (p *primedCodec) NewEncoder(w io.Writer) Encoder {
	var enc delegateEncoder
	enc.Encoder = p.codec.NewEncoder(&enc)
	enc.Writer = ioutil.Discard
	// This same input was already tested on Primer construction.
	// Therefore we will assume no error is returned.
	enc.Encode(p.types)
	enc.Writer = w
	return &enc
}

func (p *primedCodec) NewDecoder(r io.Reader) Decoder {
	var dec delegateDecoder
	dec.Decoder = p.codec.NewDecoder(&dec)
	dec.Reader = bytes.NewReader(p.data)
	var types []interface{}
	// This same input was already tested on Primer construction.
	// Therefore we will assume no error is returned.
	dec.Decode(&types)
	dec.Reader = r
	return &dec
}
