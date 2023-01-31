package codec

import (
	"bytes"
	"io"

	codecs "github.com/ugorji/go/codec"
)

// reused safely between encoders.
var codechandle codecs.BincHandle
var jsonhandle codecs.JsonHandle

func MarshalJsonIntoWriter(args interface{}, writer io.Writer) error {
	return codecs.NewEncoder(writer, &jsonhandle).Encode(args)
}
func UnmarshalJson(data string, v interface{}) error {
	d := codecs.NewDecoderString(data, &codecs.JsonHandle{})
	if err := d.Decode(v); err != nil {
		return err
	}
	return nil
}

func Marshal(args interface{}) ([]byte, error) {
	var buffer bytes.Buffer
	enc := codecs.NewEncoder(&buffer, &codechandle)

	if err := enc.Encode(args); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func MarshalIntoWriter(args interface{}, writer io.Writer) error {
	return codecs.NewEncoder(writer, &codechandle).Encode(args)
}

// reconstruct data
func Unmarshal(data []byte, v interface{}) error {
	return codecs.NewDecoder(bytes.NewBuffer(data), &codechandle).Decode(v)
}
