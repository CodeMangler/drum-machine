// Package drum implements the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

const (
	headerMetadataSize       = 40
	pascalStringMetadataSize = 1
	trackIDSize              = 4
	trackStepsSize           = 16
)

// Header represents the .splice file header.
type Header struct {
	signature     [6]byte
	contentLength uint64
	version       [32]byte
	tempo         float32
}

// parseHeader parses byte stream from an io.Reader and creates a Header structure.
func parseHeader(r io.Reader) (Header, error) {
	header := Header{}
	io.ReadFull(r, header.signature[0:])
	binary.Read(r, binary.BigEndian, &header.contentLength)
	io.ReadFull(r, header.version[0:])
	binary.Read(r, binary.LittleEndian, &header.tempo)
	return header, nil
}

// versionString returns version of the HW used to create the .splice file.
func (header Header) versionString() string {
	nullIndex := bytes.IndexByte(header.version[:], 0x00)
	return string(header.version[:nullIndex])
}

// ContentLength returns the number of bytes of Track content data in the .splice file.
func (header Header) ContentLength() uint64 {
	return uint64(header.contentLength - headerMetadataSize)
}

// String returns a string representation of the .splice file header.
func (header Header) String() string {
	return fmt.Sprintf("Saved with HW Version: %s\nTempo: %v", header.versionString(), header.tempo)
}

// Track represents data contained in a single track of a .splice drum machine file.
type Track struct {
	id    uint32
	name  PascalString
	steps [16]uint8
}

func parseTrack(r io.Reader) (Track, error) {
	track := Track{}
	binary.Read(r, binary.LittleEndian, &track.id)
	track.name, _ = parsePascalString(r)
	io.ReadFull(r, track.steps[0:])
	return track, nil
}

func parseTrackCollection(r io.Reader, bytesToRead uint64) ([]Track, error) {
	tracks := []Track{}
	bytesRead := uint64(0)
	for bytesRead < bytesToRead {
		track, _ := parseTrack(r)
		bytesRead += track.Size()
		tracks = append(tracks, track)
	}
	return tracks, nil
}

// Size returns the number of bytes taken by the Track in memory.
func (track Track) Size() uint64 {
	return uint64(trackIDSize + track.name.Size() + trackStepsSize)
}

// String returns a string representation of the Track.
func (track Track) String() string {
	trackString := fmt.Sprintf("(%d) %s\t|", track.id, fmt.Sprint(track.name))
	for index, step := range track.steps {
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

// PascalString represents a length prefixed string. Named so because of it's similarity to string representation in Pascal.
type PascalString struct {
	length uint8
	text   []byte
}

// parsePascalString parses byte stream from an io.Reader and creates a PascalString.
func parsePascalString(r io.Reader) (PascalString, error) {
	pstring := PascalString{}
	binary.Read(r, binary.LittleEndian, &pstring.length)
	pstring.text = make([]byte, pstring.length)
	io.ReadFull(r, pstring.text)
	return pstring, nil
}

// Size returns the number of bytes taken by the PascalString in memory.
func (pstring PascalString) Size() uint64 {
	return uint64(pstring.length + pascalStringMetadataSize)
}

// String returns a string representation of the PascalString.
func (pstring PascalString) String() string {
	return fmt.Sprintf(string(pstring.text[:pstring.length]))
}
