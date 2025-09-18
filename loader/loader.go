package loader

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"errors"
	"os"
	"strings"
)

// SaveCSV saves [][]string to CSV file
func SaveCSV(records [][]string, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	err = w.WriteAll(records)
	if err != nil {
		return err
	}
	w.Flush()
	return w.Error()
}

// SaveJSON saves [][]string as JSON array of objects
func SaveJSON(records [][]string, filename string) error {
	if len(records) < 1 {
		return errors.New("no data to save")
	}
	headers := records[0]
	var objs []map[string]string
	for _, row := range records[1:] {
		obj := map[string]string{}
		for i, h := range headers {
			if i < len(row) {
				obj[h] = row[i]
			}
		}
		objs = append(objs, obj)
	}

	data, err := json.MarshalIndent(objs, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// SaveFlatXML saves [][]string as flattened XML (root->Row->columns)
func SaveFlatXML(records [][]string, filename string) error {
	if len(records) < 1 {
		return errors.New("no data to save")
	}
	headers := records[0]

	type Row map[string]string
	type Rows struct {
		XMLName xml.Name `xml:"Rows"`
		Rows    []Row    `xml:"Row"`
	}

	var rowList []Row
	for _, row := range records[1:] {
		r := Row{}
		for i, h := range headers {
			if i < len(row) {
				r[h] = row[i]
			}
		}
		rowList = append(rowList, r)
	}

	dataStruct := Rows{Rows: rowList}

	data, err := xml.MarshalIndent(dataStruct, "", "  ")
	if err != nil {
		return err
	}

	// Write XML header + data
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write([]byte(xml.Header))
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	return err
}

// SaveExcelXML saves [][]string as Excel-compatible XML (.xls or .xml)
func SaveExcelXML(records [][]string, filename string) error {
	if len(records) < 1 {
		return errors.New("no data to save")
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	// Excel XML header
	header := `<?xml version="1.0"?>
<?mso-application progid="Excel.Sheet"?>
<Workbook xmlns="urn:schemas-microsoft-com:office:spreadsheet"
 xmlns:o="urn:schemas-microsoft-com:office:office"
 xmlns:x="urn:schemas-microsoft-com:office:excel"
 xmlns:ss="urn:schemas-microsoft-com:office:spreadsheet"
 xmlns:html="http://www.w3.org/TR/REC-html40">
 <Worksheet ss:Name="Sheet1">
  <Table>
`
	footer := `  </Table>
 </Worksheet>
</Workbook>`

	_, err = f.WriteString(header)
	if err != nil {
		return err
	}

	for _, row := range records {
		_, err = f.WriteString("   <Row>\n")
		if err != nil {
			return err
		}
		for _, cell := range row {
			_, err = f.WriteString("    <Cell><Data ss:Type=\"String\">" + escapeXML(cell) + "</Data></Cell>\n")
			if err != nil {
				return err
			}
		}
		_, err = f.WriteString("   </Row>\n")
		if err != nil {
			return err
		}
	}

	_, err = f.WriteString(footer)
	return err
}

// escapeXML escapes &, <, >, ' and " for XML content
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}
