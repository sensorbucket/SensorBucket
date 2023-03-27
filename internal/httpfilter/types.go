package httpfilter

var _ StringConverter = (*Bytes)(nil)

type Bytes []byte

func (Bytes) FromString(str string) (any, error) {
	return []byte(str), nil
}
