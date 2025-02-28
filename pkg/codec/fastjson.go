package codec

import (
	"io"

	"github.com/json-iterator/go"
	"github.com/pkg/errors"
)

type fastJsonCodec struct{}

func NewFastJsonCodec() Codec {
	return &fastJsonCodec{}
}

func (_ *fastJsonCodec) Unmarshal(data []byte, v interface{}) error {
	err := jsoniter.Unmarshal(data, v)
	return errors.WithStack(err)
}

func (_ *fastJsonCodec) Marshal(v interface{}) ([]byte, error) {
	if reader, ok := v.(io.Reader); ok {
		return io.ReadAll(reader)
	}
	data, err := jsoniter.Marshal(v)
	return data, errors.WithStack(err)
}
