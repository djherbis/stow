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
// NewEncoder/NewDecoder. The order that types are listed (even through recursive encoding of a type)
// of the passed sample types must be consistent across re-loads. This means that modifying
// the definition of sample types may not be supported. Also, avoid introducing types via
// map keys/values as map iteration isn't consistent. Introducing new types during encoding
// will prevent you from adding new types to the Primed type-set, and also will include
// the type definiftion over-head in thier output.
// Warning, PrimedCodec should be used consistently (for reading & writing). It
// won't be able to read data written by unprimed encoders, and data written by it
// won't be able to be read by unprimed decoders.
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
