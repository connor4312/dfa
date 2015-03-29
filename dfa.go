package main

import (
	"bufio"
	"flag"
	"github.com/ajstarks/svgo"
	"log"
	"os"
)

var input = flag.String("input", "", "path to the input DFA file")
var output = flag.String("output", "", "path to save the file to")

func main() {
	flag.Parse()
	graph := graphFile()

	file, err := os.Create(*output)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	canvas := svg.New(file)
	canvas.Start(WIDTH, HEIGHT)
	graph.Start.Plot(canvas, WIDTH/2, PADDING+ENTRY_SIZE+NODE_BASE_RAD)
	canvas.End()
}

// Parses the input file and returns a Graph representing the
// DFA that it contains.
func graphFile() *Graph {
	file, err := os.Open(*input)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	graph := &Graph{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		graph.Parse(scanner.Text())
	}

	log.Println("Graphing complete.")
	return graph
}
