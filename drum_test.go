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
	assert.Equal(t, uint64(60), header.contentSize())
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

func TestHeaderParserErrorHandling(t *testing.T) {
	headerBytes := []byte{'S', 'P', 'L', 'I', 'C', 'E',
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xEF,
		'0', '.', '8', '0', '8', '-', 'a', 'l', 'p', 'h', 'a', 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0xCD, 0xCC, 0xC4, 0x42, 0x00, 0x00, 0x00, 0x00}

	testCases := []struct {
		name         string
		sliceStart   int
		sliceEnd     int
		errorMessage string
	}{{"Bad Signature", 1, 52, "error when parsing header: signature mismatch"},
		{"EOF while reading a field", 0, 4, "error when parsing header signature: unexpected EOF"},
		{"EOF before beginning to read a field", 1, 15, "error when parsing header version: EOF"},
	}

	for _, test := range testCases {
		t.Logf("Test case: %s\n", test.name)
		buffer := bytes.NewBuffer(headerBytes[test.sliceStart:test.sliceEnd])
		header, error := parseHeader(buffer)
		if header != nil {
			t.Fatalf("Expected header parsing to fail. Got:%s\n", header)
		}
		assert.Equal(t, test.errorMessage, error.Error())
	}
}

func TestTrackParsing(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{0x63, 0x00, 0x00, 0x00,
		0x09, 'L', 'o', 'w', ' ', 'C', 'o', 'n', 'g', 'a',
		0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00})
	track, _ := parseTrack(buffer)

	assert.Equal(t, uint32(99), track.id)
	assert.Equal(t, 9, int(track.name.length))
	assert.Equal(t, "Low Conga", string(track.name.text))
	assert.Equal(t, [16]uint8{0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00}, track.steps)
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
	assert.Equal(t, uint32(255), tracks[0].id)
	assert.Equal(t, 9, int(tracks[0].name.length))
	assert.Equal(t, "Low Conga", string(tracks[0].name.text))
	assert.Equal(t, [16]uint8{0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00}, tracks[0].steps)
	assert.Equal(t, uint32(99), tracks[1].id)
	assert.Equal(t, 7, int(tracks[1].name.length))
	assert.Equal(t, "Maracas", string(tracks[1].name.text))
	assert.Equal(t, [16]uint8{0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00}, tracks[1].steps)
}

func TestTrackSize(t *testing.T) {
	track := Track{id: 220,
		name:  PascalString{length: 9, text: []byte{'L', 'o', 'w', ' ', 'C', 'o', 'n', 'g', 'a'}},
		steps: [16]uint8{0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00}}
	assert.Equal(t, uint64(30), track.size())
}

func TestTrackStringRepresentation(t *testing.T) {
	track := Track{id: 220,
		name:  PascalString{length: 9, text: []byte{'L', 'o', 'w', ' ', 'C', 'o', 'n', 'g', 'a'}},
		steps: [16]uint8{0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00}}
	expectedStringRepresentation := "(220) Low Conga\t|---x|----|---x|----|"
	assert.Equal(t, expectedStringRepresentation, fmt.Sprint(track))
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

	assert.Equal(t, uint64(12), pascalString.size())
}

func TestPascalStringStringRepresentation(t *testing.T) {
	pascalString := PascalString{length: 11, text: []byte{'T', 'e', 's', 't', ' ', 'S', 't', 'r', 'i', 'n', 'g'}}

	assert.Equal(t, "Test String", fmt.Sprint(pascalString))
}
