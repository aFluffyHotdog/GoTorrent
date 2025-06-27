package peers

import (
	"encoding/binary"
	"fmt"
	"io"
)

type messageID uint8

const (
	MsgChoke         messageID = 0 // peer is not yet ready
	MsgUnchoke       messageID = 1 // peer is now ready
	MsgInterested    messageID = 2 // downloader wants the file the client has
	MsgNotInterested messageID = 3
	MsgHave          messageID = 4
	MsgBitfield      messageID = 5
	MsgRequest       messageID = 6
	MsgPiece         messageID = 7
	MsgCancel        messageID = 8
)

// Message stores ID and payload of a message
type Message struct {
	ID      messageID
	Payload []byte
}

// Serializes a message into RAW bytes
// <length prefix><message ID><payload>
func (m *Message) SerializeMessage() []byte {
	if m == nil {
		return make([]byte, 4)
	}
	length := uint32(1 + len(m.Payload)) // message flag + content
	//fmt.Println("Serializing message of length: ", length)
	buf := make([]byte, 4+length) // len bit + message flag + content
	binary.BigEndian.PutUint32(buf[0:4], length)
	buf[4] = byte(m.ID)
	copy(buf[5:], m.Payload)
	return buf
}

// FormatRequest creates a REQUEST message
func FormatRequest(index, begin, length int) *Message {
	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[0:4], uint32(index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(begin))
	binary.BigEndian.PutUint32(payload[8:12], uint32(length))
	return &Message{ID: MsgRequest, Payload: payload}
}

// FormatHave creates a HAVE message
func FormatHave(index int) *Message {
	payload := make([]byte, 4)
	binary.BigEndian.PutUint32(payload, uint32(index))
	return &Message{ID: MsgHave, Payload: payload}
}

// Returns the integer payload of the HAVE message
func ParseHave(m *Message) (int, error) {
	if m.ID != MsgHave {
		return 0, fmt.Errorf("expected HAVE (ID %d), got ID %d", MsgHave, m.ID)
	}

	if len(m.Payload) != 4 { //payload must be a 32 bit integer
		return 0, fmt.Errorf("malformed payload")
	}

	return int(binary.BigEndian.Uint32(m.Payload)), nil
}

// Parses a torrent PIECE message and copies the contents into a buffer, returns the number of bytes read
// Message structure:  <len=0009+X><id=7><index><begin><block>
func ParsePiece(m *Message, buf []byte, index int) (int, error) {
	if m.ID != MsgPiece {
		return 0, fmt.Errorf("expected HAVE (ID %d), got ID %d", MsgHave, m.ID)
	}
	rcvIndex := int(binary.BigEndian.Uint32(m.Payload[0:4]))
	if rcvIndex != index {
		return 0, fmt.Errorf("expected piece no. (%d), got %d", index, rcvIndex)
	}

	begin := int(binary.BigEndian.Uint32(m.Payload[4:8]))
	if begin >= len(buf) {
		return 0, fmt.Errorf("begin offset bigger than buffer size. %d >= %d", begin, len(buf))
	}

	block := m.Payload[8:]
	if begin+len(block) > len(buf) {
		return 0, fmt.Errorf("data too long [%d] for offset %d with length %d", len(block), begin, len(buf))
	}
	copy(buf[begin:], block)

	return len(block), nil
}

// Converts message's raw bytes into a message object
func Read(r io.Reader) (*Message, error) {
	lenBuf := make([]byte, 4)
	_, err := io.ReadFull(r, lenBuf)
	if err != nil {
		return nil, fmt.Errorf("malformed message")
	}

	length := binary.BigEndian.Uint32(lenBuf)

	// Keep alive message, lets the other peer know "Hey! I'm still here!"
	if length == 0 {
		return nil, nil
	}

	messageBuf := make([]byte, length)
	_, err = io.ReadFull(r, messageBuf)
	if err != nil {
		return nil, err
	}

	m := Message{
		ID:      messageID(messageBuf[0]),
		Payload: messageBuf[1:],
	}
	return &m, nil
}
