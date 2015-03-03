package drum

import (
	"bufio"
	"fmt"
	"os"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
// TODO: implement
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

	p.Header, _ = parseHeader(bufferedReader)
	p.Tracks, _ = parseTrackCollection(bufferedReader, p.Header.ContentLength-40)
	fmt.Println(fmt.Sprint(p.Tracks[0]))
	return p, nil
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
// TODO: implement
type Pattern struct {
	Header FileHeader
	Tracks []Track
}

func (pattern Pattern) String() string {
	patternString := fmt.Sprintf("%s\n", pattern.Header)
	for _, track := range pattern.Tracks {
		patternString += fmt.Sprintf("%s\n", track)
	}
	return patternString
}
