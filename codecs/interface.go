package codecs

// Codec defines the interface for encoding implementations.
type Codec interface {
	// Marshal encodes the given data and returns it as a byte slice.
	Marshal(data any) ([]byte, error)

	// Unmarshal decodes the given byte slice into the destination pointer.
	Unmarshal(data []byte, dest any) error
}
