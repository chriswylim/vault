package jsonutil

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"

	"github.com/hashicorp/vault/helper/compressutil"
)

// Encodes/Marshals the given object into JSON
func EncodeJSON(in interface{}) ([]byte, error) {
	if in == nil {
		return nil, fmt.Errorf("input for encoding is nil")
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(in); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Decodes/Unmarshals the given JSON into a desired object
func DecodeJSON(data []byte, out interface{}) error {
	if data == nil {
		return fmt.Errorf("'data' being decoded is nil")
	}
	if out == nil {
		return fmt.Errorf("output parameter 'out' is nil")
	}

	return DecodeJSONFromReader(bytes.NewReader(data), out)
}

// Decodes/Unmarshals the given io.Reader pointing to a JSON, into a desired object
func DecodeJSONFromReader(r io.Reader, out interface{}) error {
	if r == nil {
		return fmt.Errorf("'io.Reader' being decoded is nil")
	}
	if out == nil {
		return fmt.Errorf("output parameter 'out' is nil")
	}

	dec := json.NewDecoder(r)

	// While decoding JSON values, intepret the integer values as `json.Number`s instead of `float64`.
	dec.UseNumber()

	// Since 'out' is an interface representing a pointer, pass it to the decoder without an '&'
	return dec.Decode(out)
}

// DecompressAndDecodeJSON tries to decompress the given data. The call to
// decompress, fails if the content was not compressed in the first place,
// which is identified by a canary byte before the compressed data. If the data
// is not compressed, it is JSON decoded directly. Otherwise the decompressed
// data will be JSON decoded.
func DecompressAndDecodeJSON(dataBytes []byte, out interface{}) error {
	if dataBytes == nil || len(dataBytes) == 0 {
		return fmt.Errorf("'dataBytes' being decoded is invalid")
	}
	if out == nil {
		return fmt.Errorf("output parameter 'out' is nil")
	}

	// Decompress the dataBytes using Gzip format. Decompression when using Gzip
	// is agnostic of the compression levels used during compression.
	decompressedBytes, unencrypted, err :=
		compressutil.Decompress(dataBytes, &compressutil.CompressionConfig{
			Type: compressutil.CompressionTypeGzip,
		})
	if err != nil {
		return fmt.Errorf("failed to decompress JSON: err: %v", err)
	}

	// If the dataBytes supplied failed to contain the compression canary, it
	// can be inferred that it was not compressed in the first place.  Try
	// to decode it.
	if unencrypted {
		return DecodeJSON(dataBytes, out)
	}

	if decompressedBytes == nil || len(decompressedBytes) == 0 {
		return fmt.Errorf("decompressed data being decoded is invalid")
	}

	// JSON decode the decompressed data
	return DecodeJSON(decompressedBytes, out)
}

// EncodeJSONAndCompress encodes the given input into JSON and compresses the
// encoded value using Gzip format (BestCompression level). A canary byte is
// placed at the beginning of the returned bytes for the logic in decompression
// method to identify compressed input.
func EncodeJSONAndCompress(in interface{}) ([]byte, error) {
	if in == nil {
		return nil, fmt.Errorf("input for encoding is nil")
	}

	// First JSON encode the given input
	encodedBytes, err := EncodeJSON(in)
	if err != nil {
		return nil, err
	}

	// For compression, use Gzip format with 'BestCompression' level.
	return compressutil.Compress(encodedBytes, &compressutil.CompressionConfig{
		Type:                 compressutil.CompressionTypeGzip,
		GzipCompressionLevel: gzip.BestCompression,
	})
}
