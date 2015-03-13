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
	}{{"Bad Signature", 1, 52, "error while parsing header: signature mismatch"},
		{"EOF while parsing a field", 0, 4, "error while parsing header signature: unexpected EOF"},
		{"EOF before beginning to parse a field", 1, 15, "error while parsing header version: EOF"}}

	for _, test := range testCases {
		t.Logf("Test case: %s\n", test.name)
		buffer := bytes.NewBuffer(headerBytes[test.sliceStart:test.sliceEnd])
		header, error := parseHeader(buffer)
		if header != nil {
			t.Fatalf("Expected header parsing to fail. Got:\t%s\n", header)
		}
		assert.Equal(t, test.errorMessage, error.Error())
	}
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

func TestTrackParserErrorHandling(t *testing.T) {
	trackBytes := []byte{0x63, 0x00, 0x00, 0x00,
		0x09, 'L', 'o', 'w', ' ', 'C', 'o', 'n', 'g', 'a',
		0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00}

	testCases := []struct {
		name         string
		sliceStart   int
		sliceEnd     int
		errorMessage string
	}{{"EOF while parsing track id", 0, 3, "error while parsing track id: unexpected EOF"},
		{"EOF while parsing track steps", 0, 16, "error while parsing track steps: unexpected EOF"},
		{"EOF before beginning to parse a field", 0, 14, "error while parsing track steps: EOF"}}

	for _, test := range testCases {
		t.Logf("Test case: %s\n", test.name)
		buffer := bytes.NewBuffer(trackBytes[test.sliceStart:test.sliceEnd])
		track, error := parseTrack(buffer)
		if track != nil {
			t.Fatalf("Expected track parsing to fail. Got:\t%s\n", track)
		}
		assert.Equal(t, test.errorMessage, error.Error())
	}
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

func TestTrackCollectionParserErrorHandling(t *testing.T) {
	content := []byte{0xFF, 0x00, 0x00, 0x00,
		0x09, 'L', 'o', 'w', ' ', 'C', 'o', 'n', 'g', 'a',
		0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00,
		0x63, 0x00, 0x00, 0x00,
		0x07, 'M', 'a', 'r', 'a', 'c', 'a', 's',
		0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01}

	buffer := bytes.NewBuffer(content)
	tracks, err := parseTrackCollection(buffer, uint64(len(content)))

	if len(tracks) != 1 {
		t.Errorf("Exactly one track should have been parsed correctly. Got: %d", len(tracks))
	}
	assert.Equal(t, "error while parsing track collection: error while parsing track steps: unexpected EOF", err.Error())
}

func TestTrackSize(t *testing.T) {
	track := Track{id: 220,
		name:  &PascalString{length: 9, text: []byte{'L', 'o', 'w', ' ', 'C', 'o', 'n', 'g', 'a'}},
		steps: [16]uint8{0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00}}
	assert.Equal(t, uint64(30), track.size())
}

func TestTrackStringRepresentation(t *testing.T) {
	track := Track{id: 220,
		name:  &PascalString{length: 9, text: []byte{'L', 'o', 'w', ' ', 'C', 'o', 'n', 'g', 'a'}},
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

func TestPascalStringParserErrorHandling(t *testing.T) {
	stringBytes := []byte{11, 'T', 'e', 's', 't', ' ', 'S', 't', 'r', 'i', 'n', 'g'}

	testCases := []struct {
		name         string
		sliceStart   int
		sliceEnd     int
		errorMessage string
	}{{"EOF while parsing pascal string length", 0, 0, "error while parsing pascal string length: EOF"},
		{"EOF while parsing pascal string text", 0, 4, "error while parsing pascal string text: unexpected EOF"},
		{"EOF before beginning to parse a field", 0, 1, "error while parsing pascal string text: EOF"}}

	for _, test := range testCases {
		t.Logf("Test case: %s\n", test.name)
		buffer := bytes.NewBuffer(stringBytes[test.sliceStart:test.sliceEnd])
		pstring, error := parsePascalString(buffer)
		if pstring != nil {
			t.Fatalf("Expected pascal string parsing to fail. Got:%s\n", pstring)
		}
		assert.Equal(t, test.errorMessage, error.Error())
	}
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
