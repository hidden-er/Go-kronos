package utils

import (
	"bytes"
	"encoding/binary"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing"
	"sort"
)

func MapToSlice(data map[int][]string) [][]string {
	// 获取 map 的所有键并排序
	keys := make([]int, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Ints(keys) // 按照键从小到大排序

	// 按排序后的键依次将值添加到切片中
	result := make([][]string, 0, len(keys))
	for _, k := range keys {
		result = append(result, data[k])
	}

	return result
}

func MessageEncap(m [][]byte) []byte {
	var buf bytes.Buffer
	for j := 0; j < len(m); j++ {
		buf.Write(m[j])
	}
	return buf.Bytes()
}

func PointToBytes(P kyber.Point) []byte {
	B, _ := P.MarshalBinary()
	return B
}

func BytesToPoint(B []byte) kyber.Point {
	P := pairing.NewSuiteBn256().Point()
	P.UnmarshalBinary(B)
	return P
}

//Uint32ToBytes convert uint32 to bytes
func Uint32ToBytes(n uint32) []byte {
	bytebuf := bytes.NewBuffer([]byte{})
	binary.Write(bytebuf, binary.BigEndian, n)
	return bytebuf.Bytes()
}

//BytesToUint32 convert bytes to uint32
func BytesToUint32(byt []byte) uint32 {
	bytebuff := bytes.NewBuffer(byt)
	var data uint32
	binary.Read(bytebuff, binary.BigEndian, &data)
	return data
}

//BytesToInt convert bytes to int
func BytesToInt(byt []byte) int {
	bytebuff := bytes.NewBuffer(byt)
	var data uint32
	binary.Read(bytebuff, binary.BigEndian, &data)
	return int(data)
}

//IntToBytes convert int to bytes
func IntToBytes(n int) []byte {
	data := uint32(n)
	bytebuf := bytes.NewBuffer([]byte{})
	binary.Write(bytebuf, binary.BigEndian, data)
	return bytebuf.Bytes()
}

//Uint32sToBytes convert uint32s to bytes
func Uint32sToBytes(ns []uint32) []byte {
	bytebuf := bytes.NewBuffer([]byte{})
	for _, n := range ns {
		binary.Write(bytebuf, binary.BigEndian, n)
	}
	return bytebuf.Bytes()
}

//BytesToUint32s convert bytes to uint32s
func BytesToUint32s(byt []byte) []uint32 {
	bytebuff := bytes.NewBuffer(byt)
	data := make([]uint32, len(byt)/4)
	binary.Read(bytebuff, binary.BigEndian, &data)
	return data
}
