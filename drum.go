// Package drum implements the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// Header represents the .splice file header
type Header struct {
	Signature     [6]byte
	ContentLength uint64
	VersionString [32]byte
	Tempo         float32
}

// VersionStringText returns version of the HW used to create the .splice file
func (header Header) VersionStringText() string {
	nullIndex := bytes.IndexByte(header.VersionString[:], 0x00)
	return string(header.VersionString[:nullIndex])
}

func (header Header) String() string {
	return fmt.Sprintf("Saved with HW Version: %s\nTempo: %v", header.VersionStringText(), header.Tempo)
}

func parseHeader(r io.Reader) (Header, error) {
	header := Header{}
	io.ReadFull(r, header.Signature[0:])
	binary.Read(r, binary.BigEndian, &header.ContentLength)
	io.ReadFull(r, header.VersionString[0:])
	binary.Read(r, binary.LittleEndian, &header.Tempo)
	return header, nil
}

// PascalString represents a length prefixed string. Named so because of it's similarity to string representation in Pascal.
type PascalString struct {
	Length uint8
	Text   []byte
}

func (pascalString PascalString) String() string {
	return fmt.Sprintf(string(pascalString.Text[:pascalString.Length]))
}

func parsePascalString(r io.Reader) (PascalString, error) {
	pascalString := PascalString{}
	binary.Read(r, binary.LittleEndian, &pascalString.Length)
	pascalString.Text = make([]byte, pascalString.Length)
	io.ReadFull(r, pascalString.Text)
	return pascalString, nil
}

// Track represents data contained in a single track of a .splice drum machine file
type Track struct {
	ID    uint32
	Name  PascalString
	Steps [16]uint8
}

func (track Track) String() string {
	trackString := fmt.Sprintf("(%d) %s\t|", track.ID, fmt.Sprint(track.Name))
	for index, step := range track.Steps {
		if step == 0x00 {
			trackString += "-"
		} else {
			trackString += "x"
		}
		if (index+1)%4 == 0 {
			trackString += "|"
		}
	}

	return trackString
}

func parseTrack(r io.Reader) (Track, error) {
	track := Track{}
	binary.Read(r, binary.LittleEndian, &track.ID)
	track.Name, _ = parsePascalString(r)
	io.ReadFull(r, track.Steps[0:])
	return track, nil
}

func parseTrackCollection(r io.Reader, bytesToRead uint64) ([]Track, error) {
	tracks := []Track{}
	bytesRead := uint64(0)
	for bytesRead < bytesToRead {
		track, _ := parseTrack(r)
		thisTrack := uint64(4 + 1 + track.Name.Length + 16)
		bytesRead += thisTrack
		tracks = append(tracks, track)
	}
	return tracks, nil
}
