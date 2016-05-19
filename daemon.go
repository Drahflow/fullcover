package main

import (
	"net/http"
	"sync"
	"bufio"
	"log"
	"fmt"
	"io"
	"strings"
)

// sources holds all reported sources
var sources = make(map[string]string)

type block struct {
	startLine int
	startCol int
	endLine int
	endCol int
	numStmt int
	count int
}

// counts, key is [source][startLine][startCol]
var counts map[string]map[int]map[int]*block
var countsLock sync.Mutex

func runDaemon() {
	http.HandleFunc("/coverage", collectCoverage)
	http.HandleFunc("/", handleReporting)

	http.ListenAndServe(*connection, nil)
}

func collectCoverage(w http.ResponseWriter, r *http.Request) {
	reader := bufio.NewReader(r.Body)

	for {
		first, err := reader.ReadByte()

		if err != nil {
			break
		}

		switch first {
		case 'F':
			filename := readNetstring(reader)
			source := readNetstring(reader)

			countsLock.Lock()
			sources[filename] = source
			countsLock.Unlock()

		case 'C':
			collectBlock(reader, 1)

		case 'B':
			collectBlock(reader, 0)

		default:
			log.Fatalf("Invalid coverage report header: %d", first)
		}
	}
}

func collectBlock(reader *bufio.Reader, delta int) {
	filename := readNetstring(reader)
	startLine := readInt(reader)
	startCol := readInt(reader)
	endLine := readInt(reader)
	endCol := readInt(reader)
	numStmt := readInt(reader)

	countsLock.Lock()
	if counts == nil {
		counts = make(map[string]map[int]map[int]*block)
	}

	if counts[filename] == nil {
		counts[filename] = make(map[int]map[int]*block)
	}

	if counts[filename][startLine] == nil {
		counts[filename][startLine] = make(map[int]*block)
	}

	if counts[filename][startLine][startCol] == nil {
		counts[filename][startLine][startCol] = &block{
			startLine: startLine,
			startCol: startCol,
			endLine: endLine,
			endCol: endCol,
			numStmt: numStmt,
			count: 0,
		}
	}

	counts[filename][startLine][startCol].count += delta

	countsLock.Unlock()
}

func readInt(r *bufio.Reader) int {
	str, err := r.ReadString(':')
	if err != nil {
		log.Fatalf("could not parse cover report: %v", err)
	}

	var result int
	count, err := fmt.Sscanf(str, "%d", &result)
	if err != nil {
		log.Fatalf("could not parse cover report: %v", err)
	}

	if count != 1 {
		log.Fatalf("could not parse cover report, no count")
	}

	return result
}

func readNetstring(r *bufio.Reader) string {
	length := readInt(r)
	resultBuf := make([]byte, length)
	count, err := io.ReadFull(r, resultBuf)
	if err != nil {
		log.Fatalf("could not parse cover report: %v", err)
	}

	if count != len(resultBuf) {
		log.Fatalf("could not parse cover report, ReadFull returned wrong count")
	}

	return string(resultBuf)
}

func handleReporting(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		handleIndex(w, r)
		return
	}

	_, ok := sources[r.URL.Path[1:]]
	if ok {
		handleSource(w, r, r.URL.Path[1:])
		return
	}

	http.NotFound(w, r)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `
<html><head>
</head><body>
  <ul>
`)
	for filename := range sources {
		totalStmt := 0
		coveredStmt := 0

		for _, lineCounts := range counts[filename] {
			for _, block := range lineCounts {
				totalStmt += block.numStmt

				if block.count > 0 {
					coveredStmt += block.numStmt
				}
			}
		}

		if totalStmt > 0 {
			fmt.Fprintf(w, `
	<li><a href="%s">%s</a> (%3.2f%% %d/%d)</li>
`, filename, filename, float32(coveredStmt) / float32(totalStmt) * 100, coveredStmt, totalStmt)
		} else {
			fmt.Fprintf(w, `
	<li><a href="%s">%s</a> (no statements)</li>
`, filename, filename)
		}
	}

	fmt.Fprintf(w, `
  </ul>
</body></html>
`)
}

func handleSource(w http.ResponseWriter, r *http.Request, filename string) {
	fmt.Fprintf(w, `
<html><head>
</head><body style="background-color: black; color: white;">
<pre>%s`, changeColor(0))

	lines := strings.Split(sources[filename], "\n")
	n := make([][]int, len(lines))

	for i, line := range lines {
		n[i] = make([]int, len(line))

		for j := 0; j < len(line); j++ {
			n[i][j] = -1
		}
	}

	for _, lineData := range counts[filename] {
		for _, data := range lineData {
			y := data.startLine - 1
			x := data.startCol - 1

			for {
				n[y][x] = data.count
				x++

				for x >= len(n[y]) {
					x = 0
					y++

					if y >= len(n) {
						break
					}
				}

				if y > data.endLine - 1 || ( y == data.endLine - 1 && x > data.endCol - 1 ) {
					break
				}
			}
		}
	}

	lastCount := 0
	for y, lineCount := range n {
		for x, c := range lineCount {
			if c != lastCount {
				fmt.Fprintf(w, "%s", changeColor(c))
				lastCount = c
			}

			fmt.Fprintf(w, "%c", lines[y][x])
		}
		fmt.Fprintf(w, "\n")
	}

	fmt.Fprintf(w, `
</span></pre>
</body></html>
`)
}

func changeColor(c int) string {
	switch {
	case c == -1:
		return `</span><span style="color: #808080">`
	case c == 0:
		return `</span><span style="color: #ff0000" title="0">`
	default:
		green := 255
		blue := 0

		return fmt.Sprintf(`</span><span style="color: #00%02x%02x" title="%d">`,
			green, blue, c)
	}
}
