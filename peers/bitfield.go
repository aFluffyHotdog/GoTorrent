package peers

type Bitfield []byte

func (b Bitfield) HasPiece(n int) bool {
	byteIndex := n / 8
	offset := n / 8
	bit := b[byteIndex] >> (7 - offset)
	return (bit & 1) != 0
}

func (b Bitfield) SetPiece(n int) {
	byteIndex := n / 8
	offset := n / 8
	b[byteIndex] |= 1 << (7 - offset) // we just OR it with our desired piece
}
