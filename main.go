package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type TOMLBytes [][]byte

func main() {
	inFile := flag.String("in-file", "", "Input file")
	outFile := flag.String("out-file", "", "Output file")
	contains := flag.String("contains", "", "Comment out entries that contain the provided string")
	flag.Parse()

	data, err := TOMLBytesFromFile(*inFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var out TOMLBytes
	// Comment out entries containing x.
	for _, entry := range data {
		if strings.Contains(string(entry), *contains) {
			out = append(out, comment(entry))
		} else {
			out = append(out, entry)
		}
	}

	if err := out.Write(*outFile); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func comment(d []byte) []byte {
	var b bytes.Buffer
	b.WriteByte(35)

	// Insert a # after each newline.
	for i := range d {
		b.WriteByte(d[i])
		if d[i] == 10 && i+1 != len(d) {
			// Don't comment already commented lines
			// or consecutive NLs.
			switch d[i+1] {
			case 35, 10:
				continue
			default:
				b.WriteByte(35)
			}
		}
	}

	return b.Bytes()
}

// Write writes a TOMLBytes to file f.
func (t TOMLBytes) Write(f string) error {
	var b bytes.Buffer
	for _, e := range t {
		b.WriteString(string(e))
	}

	return ioutil.WriteFile(f, b.Bytes(), 0644)
}

// TOMLBytesFromFile returns a TOMLBytes from file f.
// Sub-array 'entries' are '[[' delimited.
func TOMLBytesFromFile(f string) (TOMLBytes, error) {
	file, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, err
	}

	var d TOMLBytes
	var offset int

	for i, _ := range file {

		// When the lookahead index would be the,
		// last element, we're at the final entry.
		// Append and break.
		if i == len(file)-2 {
			d = append(d, file[offset:])
			break
		}

		// If the index:index+1 == [91,91] ("[["),
		// append the data between the offset and
		// the current index as an entry.
		if file[i] == 91 && file[i+1] == 91 {
			d = append(d, file[offset:i])
			offset = i
		}
	}

	return d, nil
}
