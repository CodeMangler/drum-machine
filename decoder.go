package drum

import (
	"bufio"
	"fmt"
	"os"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	p := &Pattern{}
	file, error := os.Open(path)
	if error != nil {
		panic(error)
	}

	defer func() {
		if error := file.Close(); error != nil {
			panic(error)
		}
	}()

	bufferedReader := bufio.NewReader(file)

	p.header, _ = parseHeader(bufferedReader)
	p.tracks, _ = parseTrackCollection(bufferedReader, p.header.contentSize())
	return p, nil
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	header Header
	tracks []Track
}

func (pattern Pattern) String() string {
	patternString := fmt.Sprintf("%s\n", pattern.header)
	for _, track := range pattern.tracks {
		patternString += fmt.Sprintf("%s\n", track)
	}
	return patternString
}
