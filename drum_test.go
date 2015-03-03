package drum

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHeaderParsing(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{'S', 'P', 'L', 'I', 'C', 'E',
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xEF,
		'0', '.', '8', '0', '8', '-', 'a', 'l', 'p', 'h', 'a', 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0xCD, 0xCC, 0xC4, 0x42, 0x00, 0x00, 0x00, 0x00})
	header, _ := parseHeader(buffer)

	assert.Equal(t, "SPLICE", string(header.signature[:]))
	assert.Equal(t, uint64(239), header.contentLength)
	assert.Equal(t, "0.808-alpha", string(header.version[:11]))
	assert.Equal(t, 98.4, header.tempo)
}

func TestHeaderVersionString(t *testing.T) {
	header := Header{version: [32]byte{'0', '.', '9', '0', '9', '-', 'a', 'l', 'p', 'h', 'a'}}

	assert.Equal(t, "0.909-alpha", header.versionString())
}

func TestHeaderContentLengthExcludesHeaderSize(t *testing.T) {
	header := Header{signature: [6]byte{'S', 'P', 'L', 'I', 'C', 'E'},
		contentLength: 100,
		version:       [32]byte{'0', '.', '9', '0', '9', '-', 'a', 'l', 'p', 'h', 'a'},
		tempo:         78.5}
	assert.Equal(t, uint64(60), header.ContentLength())
}

func TestHeaderStringRepresentation(t *testing.T) {
	header := Header{signature: [6]byte{'S', 'P', 'L', 'I', 'C', 'E'},
		contentLength: 100,
		version:       [32]byte{'0', '.', '9', '0', '9', '-', 'a', 'l', 'p', 'h', 'a'},
		tempo:         78.5}
	expectedStringRepresentation := `Saved with HW Version: 0.909-alpha
Tempo: 78.5`
	assert.Equal(t, expectedStringRepresentation, fmt.Sprint(header))
}

func TestTrackParsing(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{0x63, 0x00, 0x00, 0x00,
		0x09, 'L', 'o', 'w', ' ', 'C', 'o', 'n', 'g', 'a',
		0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00})
	track, _ := parseTrack(buffer)

	assert.Equal(t, uint32(99), track.ID)
	assert.Equal(t, 9, int(track.Name.length))
	assert.Equal(t, "Low Conga", string(track.Name.text))
	assert.Equal(t, [16]uint8{0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00}, track.Steps)
}

func TestTrackStringRepresentation(t *testing.T) {
	track := Track{ID: 220,
		Name:  PascalString{length: 9, text: []byte{'L', 'o', 'w', ' ', 'C', 'o', 'n', 'g', 'a'}},
		Steps: [16]uint8{0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00}}
	expectedStringRepresentation := "(220) Low Conga\t|---x|----|---x|----|"
	assert.Equal(t, expectedStringRepresentation, fmt.Sprint(track))
}

func TestTrackCollectionParsing(t *testing.T) {
	content := []byte{0xFF, 0x00, 0x00, 0x00,
		0x09, 'L', 'o', 'w', ' ', 'C', 'o', 'n', 'g', 'a',
		0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00,
		0x63, 0x00, 0x00, 0x00,
		0x07, 'M', 'a', 'r', 'a', 'c', 'a', 's',
		0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00}
	buffer := bytes.NewBuffer(content)
	tracks, _ := parseTrackCollection(buffer, uint64(len(content)))

	assert.Equal(t, 2, len(tracks))
	assert.Equal(t, uint32(255), tracks[0].ID)
	assert.Equal(t, 9, int(tracks[0].Name.length))
	assert.Equal(t, "Low Conga", string(tracks[0].Name.text))
	assert.Equal(t, [16]uint8{0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00}, tracks[0].Steps)
	assert.Equal(t, uint32(99), tracks[1].ID)
	assert.Equal(t, 7, int(tracks[1].Name.length))
	assert.Equal(t, "Maracas", string(tracks[1].Name.text))
	assert.Equal(t, [16]uint8{0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00}, tracks[1].Steps)
}

func TestPascalStringParsing(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{11, 'T', 'e', 's', 't', ' ', 'S', 't', 'r', 'i', 'n', 'g'})
	pascalString, _ := parsePascalString(buffer)

	assert.Equal(t, 11, int(pascalString.length))
	assert.Equal(t, "Test String", string(pascalString.text))
}

func TestPascalStringSize(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{11, 'T', 'e', 's', 't', ' ', 'S', 't', 'r', 'i', 'n', 'g'})
	pascalString, _ := parsePascalString(buffer)

	assert.Equal(t, uint64(12), pascalString.Size())
}

func TestPascalStringStringRepresentation(t *testing.T) {
	pascalString := PascalString{length: 11, text: []byte{'T', 'e', 's', 't', ' ', 'S', 't', 'r', 'i', 'n', 'g'}}

	assert.Equal(t, "Test String", fmt.Sprint(pascalString))
}
