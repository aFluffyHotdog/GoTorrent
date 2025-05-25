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

func (m *Message) SerializeMessage() []byte {
	length := uint32(1 + len(m.Payload)) // message flag + content
	buf := make([]byte, length)          // len bit + message flag + content
	binary.BigEndian.PutUint32(buf[0:4], length)
	buf[4] = byte(m.ID)
	copy(buf[5:], m.Payload)
	return buf
}

func ReadMessage(r io.Reader) (*Message, error) {
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
