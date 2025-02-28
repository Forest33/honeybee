package codec

type Decoder interface {
	Unmarshal(data []byte, v interface{}) error
}

type Encoder interface {
	Marshal(v interface{}) ([]byte, error)
}

type Codec interface {
	Decoder
	Encoder
}
