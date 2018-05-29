package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/wcharczuk/go-chart"
)

type wordOccurance struct {
	Word      string
	Occurance int64
}

var (
	helpFlag    = flag.Bool("h", false, "Displays this menu.")
	topx        = flag.Int("t", 20, "Top x number of words.")
	filesPath   = flag.String("-p", "./txtfiles", "Path to the text files.")
	filePattern = flag.String("-f", "*.txt", "Pattern to look up files.")
	noGraph     = flag.Bool("no-graph", false, "Don't creating the graph image.")
	graphPath   = flag.String("-g", "./zipf.png", "Path to use when creating the graph image")
	cliOut      = flag.Bool("-s", false, "Show results in terminal.")
)

func main() {
	flag.Parse()
	if *helpFlag {
		flag.PrintDefaults()
		os.Exit(0)
	}
	t1 := time.Now()
	// read in file as a string
	books, ok := collectFiles()
	if !ok {
		log.Fatalln("Failed to read books. See logs for more details.")
	}

	// Create the slice containing the words
	re := regexp.MustCompile("[A-Za-z']+")
	// Create a map with the words and count them
	zipfTable := make(map[string]int64)
	for _, book := range books {
		words := re.FindAllString(book, -1)
		for _, word := range words {
			if word == `'` {
				continue
			}
			zipfTable[strings.ToLower(word)]++
		}
	}

	// Push the words into something sortable
	sortableWords := make([]*wordOccurance, 0)
	for word, occurance := range zipfTable {
		sortableWords = append(sortableWords, &wordOccurance{Word: word, Occurance: occurance})
	}
	sort.Slice(sortableWords, func(i, j int) bool { return sortableWords[i].Occurance > sortableWords[j].Occurance })
	t2 := time.Since(t1)
	log.Printf("Read all the books in %s", t2)

	buf, err := drawGraph(sortableWords[0:*topx])
	if err != nil {
		log.Fatalf("Failed to create graph. Error %s\n", err)
	}

	if !*noGraph {
		// Graph it
		err = ioutil.WriteFile(*graphPath, buf.Bytes(), 0770)
		if err != nil {
			log.Fatalf("Failed to write graph file. Error: %s\n", err)
		}
	}

	if *cliOut {
		// Show it in text
		for i := 0; i < *topx; i++ {
			fmt.Printf("%s\t\t%d\n", sortableWords[i].Word, sortableWords[i].Occurance)
		}
	}
}

func collectFiles() ([]string, bool) {
	filePath := *filesPath + "/" + *filePattern
	log.Printf("Looking for files with: %s\n", filePath)

	files, err := filepath.Glob(*filesPath + "/" + *filePattern)
	if err != nil {
		log.Fatalf("Failed globbing files. Error: %s", err)
	}
	log.Printf("Reading in files: %s", files)
	if len(files) < 1 {
		return nil, false
	}
	allFiles := make([]string, len(files))
	for index, file := range files {
		fileText, err := ioutil.ReadFile(file)
		if err != nil {
			log.Printf("There was an error reading %s. Moving on.", file)
			continue
		}
		allFiles[index] = string(fileText)
	}
	return allFiles, true
}

func drawGraph(words []*wordOccurance) (*bytes.Buffer, error) {
	chartValues := make([]chart.Value, *topx)
	for index, word := range words {
		chartValues[index] = chart.Value{
			Label: word.Word,
			Value: float64(word.Occurance),
		}
	}
	wordChart := chart.BarChart{
		Title:      "Words by occurance",
		TitleStyle: chart.StyleShow(),
		Background: chart.Style{
			Padding: chart.Box{
				Top: 40,
			},
		},
		Height:   1024,
		BarWidth: 80,
		XAxis:    chart.StyleShow(),
		YAxis: chart.YAxis{
			Style: chart.StyleShow(),
			Range: &chart.ContinuousRange{
				Min: 0.0,
				Max: chartValues[0].Value,
			},
		},
		Bars: chartValues,
	}
	buffer := bytes.NewBuffer([]byte{})
	err := wordChart.Render(chart.PNG, buffer)
	if err != nil {
		return nil, err
	}
	return buffer, nil
}
