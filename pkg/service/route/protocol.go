package route

// IOutProtocol protocol message
type IOutProtocol interface {
	Marshal() ([]byte, error)
}

// BytesOutProtocol a row bytes message
type BytesOutProtocol []byte

func (m BytesOutProtocol) Marshal() ([]byte, error) {
	return m, nil
}

// String implement the Stringer interface
func (m BytesOutProtocol) String() string {
	return string(m)
}
