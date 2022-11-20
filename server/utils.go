package server

import (
	"encoding/binary"
)

const INT64_SIZE = 8

func BytesToInt64(src []byte) (val int64, size int) {
	return int64(binary.BigEndian.Uint64(src[:INT64_SIZE])), INT64_SIZE
}

func Int64ToBytes(val int64) (res []byte) {
	return binary.BigEndian.AppendUint64([]byte{}, uint64(val))
}
