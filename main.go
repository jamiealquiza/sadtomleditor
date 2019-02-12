package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"
)

type TOMLBytes [][]byte

type byteJob struct {
	b  []byte
	id int
}

func main() {
	start := time.Now()

	inFile := flag.String("in-file", "", "Input file")
	outFile := flag.String("out-file", "", "Output file")
	contains := flag.String("contains", "", "Comment out entries that contain the provided string")
	flag.Parse()

	data, err := TOMLBytesFromFile(*inFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var out = make(TOMLBytes, len(data))
	wg := sync.WaitGroup{}
	wg.Add(len(data))

	results := make(chan byteJob, len(data))

	// Comment out entries containing x.
	for i, entry := range data {
		go func(wg *sync.WaitGroup, r chan byteJob, entry []byte, id int) {
			j := byteJob{id: id}
			if strings.Contains(string(entry), *contains) {
				j.b = comment(entry)
			} else {
				j.b = entry
			}

			r <- j
			wg.Done()
		}(&wg, results, entry, i)
	}

	wg.Wait()
	close(results)

	for j := range results {
		out[j.id] = j.b
	}

	if err := out.Write(*outFile); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(time.Since(start))
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
// Sub-array 'entrries' are '[[' delimited.
func TOMLBytesFromFile(f string) (TOMLBytes, error) {
	file, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, err
	}

	var d TOMLBytes
	var offset int

	for i, _ := range file {

		// Break when the lookahead index
		// would be the last element.
		if i == len(file)-2 {
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
