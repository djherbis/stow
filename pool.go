package stow

import (
	"io"
	"sync"
)

type pooledCodec struct {
	encoderPool sync.Pool
	decoderPool sync.Pool
}

// NewPooledCodec creates a new Codec which re-uses Encoder/Decoders created by the codec.
// Warning, this is only useful when creating Encoders/Decoders is 'expensive' and when
// they support being re-used. In order to support re-use encoders/decoders must be
// equivalent. For example, a new Gob Encoder is not equivalent to a Gob Encoder
// which has encoded a non-primitive type [it has cached type info]. On the other hand, 
// a json Encoder is safe for re-use and so is a Primed Gob Encoder which has only encoded primed types
// since all primed types are cached for all encoders/decoders.
func NewPooledCodec(codec Codec) Codec {
	return &pooledCodec{
		encoderPool: sync.Pool{New: func() interface{} {
			var enc delegateEncoder
			enc.Encoder = codec.NewEncoder(&enc)
			return &enc
		}},
		decoderPool: sync.Pool{New: func() interface{} {
			var dec delegateDecoder
			dec.Decoder = codec.NewDecoder(&dec)
			return &dec
		}},
	}
}

func (p *pooledCodec) NewEncoder(w io.Writer) Encoder {
	enc := p.encoderPool.Get().(*delegateEncoder)
	enc.Writer = w
	return enc
}

func (p *pooledCodec) NewDecoder(r io.Reader) Decoder {
	dec := p.decoderPool.Get().(*delegateDecoder)
	dec.Reader = r
	return dec
}

func (p *pooledCodec) PutEncoder(enc Encoder) {
	p.encoderPool.Put(enc)
}

func (p *pooledCodec) PutDecoder(dec Decoder) {
	p.decoderPool.Put(dec)
}
