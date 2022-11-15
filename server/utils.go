package server

import (
	"encoding/binary"
)

func BytesToInt32(data []byte) (res int32) {
	return int32(binary.LittleEndian.Uint32(data))
}
