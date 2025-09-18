package main

import (
	"flag"
	"log"
	"path/filepath"
	"strings"

	"securefiletransfer/fetcher"
	"securefiletransfer/loader"
	"securefiletransfer/parser"
)

func main() {
	urlFlag := flag.String("url", "", "File URL to fetch (HTTP/SFTP/FTP supported)")
	formatFlag := flag.String("format", "csv", "Input format: csv, json, xml")
	outFlag := flag.String("out", "output.csv", "Output file path with extension (csv, json, xml, flatxml, xls)")
	flag.Parse()

	if *urlFlag == "" {
		log.Fatal("Please provide -url argument")
	}

	data, err := fetcher.Fetch(*urlFlag)
	if err != nil {
		log.Fatalf("Failed to fetch: %v", err)
	}
	log.Printf("Downloaded %d bytes", len(data))

	var records [][]string

	switch strings.ToLower(*formatFlag) {
	case "csv":
		records, err = parser.ParseCSV(data)
		if err != nil {
			log.Fatalf("Failed to parse CSV: %v", err)
		}
	case "json":
		v, err := parser.ParseJSON(data)
		if err != nil {
			log.Fatalf("Failed to parse JSON: %v", err)
		}
		records, err = parser.ConvertJSONTo2D(v)
		if err != nil {
			log.Fatalf("Failed to convert JSON to 2D: %v", err)
		}
	case "xml":
		records, err = parser.ConvertXMLTo2D(data)
		if err != nil {
			log.Fatalf("Failed to convert XML to 2D: %v", err)
		}
	default:
		log.Fatalf("Unsupported input format: %s", *formatFlag)
	}

	outFile := strings.TrimSpace(*outFlag)
	ext := strings.ToLower(filepath.Ext(outFile))

	if ext == "" {
		log.Fatalf("No file extension found in output filename: %s", outFile)
	}

	switch ext {
	case ".csv":
		err = loader.SaveCSV(records, outFile)
	case ".json":
		err = loader.SaveJSON(records, outFile)
	case ".flatxml":
		err = loader.SaveFlatXML(records, outFile)
	case ".xls", ".xml":
		err = loader.SaveExcelXML(records, outFile)
	default:
		log.Fatalf("Unsupported output file extension: %s", ext)
	}

	if err != nil {
		log.Fatalf("Failed to save output: %v", err)
	}

	log.Printf("Saved output to %s", outFile)
}
