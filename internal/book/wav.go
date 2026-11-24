package book

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
)

// PCMData returns verified 16-bit mono PCM without assuming a 44-byte input header.
func PCMData(path string) ([]byte, uint32, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, 0, e
	}
	defer f.Close()
	var h [12]byte
	if _, e = io.ReadFull(f, h[:]); e != nil || string(h[:4]) != "RIFF" || string(h[8:]) != "WAVE" {
		return nil, 0, errors.New("Invalid generated WAV")
	}
	var rate uint32
	valid := false
	for {
		var c [8]byte
		if _, e = io.ReadFull(f, c[:]); e != nil {
			return nil, 0, errors.New("Missing generated audio")
		}
		n := binary.LittleEndian.Uint32(c[4:])
		if n > 64<<20 {
			return nil, 0, errors.New("Generated audio chunk is too large")
		}
		b := make([]byte, n)
		if _, e = io.ReadFull(f, b); e != nil {
			return nil, 0, e
		}
		if n%2 != 0 {
			f.Seek(1, io.SeekCurrent)
		}
		switch string(c[:4]) {
		case "fmt ":
			if len(b) < 16 {
				return nil, 0, errors.New("Invalid audio format")
			}
			valid = binary.LittleEndian.Uint16(b) == 1 && binary.LittleEndian.Uint16(b[2:]) == 1 && binary.LittleEndian.Uint16(b[14:]) == 16
			rate = binary.LittleEndian.Uint32(b[4:])
		case "data":
			if !valid || rate < 8000 || rate > 96000 || len(b)%2 != 0 {
				return nil, 0, errors.New("Unsupported generated audio format")
			}
			return b, rate, nil
		}
	}
}
func WriteWAVHeader(f *os.File, rate uint32, n uint32) error {
	b := make([]byte, 44)
	copy(b, "RIFF")
	binary.LittleEndian.PutUint32(b[4:], 36+n)
	copy(b[8:], "WAVEfmt ")
	binary.LittleEndian.PutUint32(b[16:], 16)
	binary.LittleEndian.PutUint16(b[20:], 1)
	binary.LittleEndian.PutUint16(b[22:], 1)
	binary.LittleEndian.PutUint32(b[24:], rate)
	binary.LittleEndian.PutUint32(b[28:], rate*2)
	binary.LittleEndian.PutUint16(b[32:], 2)
	binary.LittleEndian.PutUint16(b[34:], 16)
	copy(b[36:], "data")
	binary.LittleEndian.PutUint32(b[40:], n)
	if _, e := f.Seek(0, 0); e != nil {
		return e
	}
	_, e := f.Write(b)
	return e
}
