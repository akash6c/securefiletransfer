package parser

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ParseCSV parses CSV data into [][]string
func ParseCSV(data []byte) ([][]string, error) {
	r := csv.NewReader(strings.NewReader(string(data)))
	r.LazyQuotes = true
	r.TrimLeadingSpace = true
	return r.ReadAll()
}

// ParseJSON parses JSON data into interface{}
func ParseJSON(data []byte) (interface{}, error) {
	var v interface{}
	err := json.Unmarshal(data, &v)
	return v, err
}

// ConvertJSONTo2D converts parsed JSON to [][]string
func ConvertJSONTo2D(data interface{}) ([][]string, error) {
	var arr []interface{}

	switch v := data.(type) {
	case []interface{}:
		arr = v
	case map[string]interface{}:
		arr = []interface{}{v}
	default:
		return nil, errors.New("invalid JSON structure")
	}

	if len(arr) == 0 {
		return nil, errors.New("empty JSON array")
	}

	headerMap := map[string]struct{}{}
	for _, item := range arr {
		obj, ok := item.(map[string]interface{})
		if !ok {
			return nil, errors.New("expected JSON array of objects")
		}
		for k := range obj {
			headerMap[k] = struct{}{}
		}
	}

	headers := make([]string, 0, len(headerMap))
	for h := range headerMap {
		headers = append(headers, h)
	}

	records := [][]string{headers}
	for _, item := range arr {
		obj := item.(map[string]interface{})
		row := make([]string, len(headers))
		for i, h := range headers {
			if val, ok := obj[h]; ok {
				row[i] = formatValue(val)
			} else {
				row[i] = ""
			}
		}
		records = append(records, row)
	}

	return records, nil
}

// XML node structure to help parse
type xmlNode struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:",any,attr"`
	Content string     `xml:",chardata"`
	Nodes   []xmlNode  `xml:",any"`
}

// ParseXML parses XML into slice of flat maps
func ParseXML(data []byte) ([]map[string]string, error) {
	var root xmlNode
	if err := xml.Unmarshal(data, &root); err != nil {
		return nil, err
	}

	childrenByName := map[string][]xmlNode{}
	for _, child := range root.Nodes {
		childrenByName[child.XMLName.Local] = append(childrenByName[child.XMLName.Local], child)
	}

	for _, nodes := range childrenByName {
		if len(nodes) > 1 {
			var result []map[string]string
			for _, node := range nodes {
				m := map[string]string{}
				flattenXMLNode(node, m)
				result = append(result, m)
			}
			return result, nil
		}
	}

	m := map[string]string{}
	flattenXMLNode(root, m)
	return []map[string]string{m}, nil
}

// flattenXMLNode recursively flattens XML node into map
func flattenXMLNode(node xmlNode, m map[string]string) {
	for _, attr := range node.Attrs {
		m[attr.Name.Local] = attr.Value
	}

	if len(node.Nodes) == 0 {
		content := strings.TrimSpace(node.Content)
		if content != "" {
			m[node.XMLName.Local] = content
		}
		return
	}

	for _, child := range node.Nodes {
		flattenXMLNode(child, m)
	}
}

// ConvertXMLTo2D converts XML bytes to [][]string
func ConvertXMLTo2D(data []byte) ([][]string, error) {
	maps, err := ParseXML(data)
	if err != nil {
		return nil, err
	}

	if len(maps) == 0 {
		return nil, errors.New("no XML data found")
	}

	headerMap := map[string]struct{}{}
	for _, m := range maps {
		for k := range m {
			headerMap[k] = struct{}{}
		}
	}

	headers := make([]string, 0, len(headerMap))
	for h := range headerMap {
		headers = append(headers, h)
	}

	records := [][]string{headers}
	for _, m := range maps {
		row := make([]string, len(headers))
		for i, h := range headers {
			if val, ok := m[h]; ok {
				row[i] = formatValue(val)
			} else {
				row[i] = ""
			}
		}
		records = append(records, row)
	}

	return records, nil
}

// formatValue formats values and converts recognized dates
func formatValue(val interface{}) string {
	switch v := val.(type) {
	case string:
		t, err := tryParseDate(v)
		if err == nil {
			return t.Format("2006-01-02")
		}
		return v
	case float64:
		return fmt.Sprintf("%v", v)
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}
}

// tryParseDate tries multiple layouts to parse date strings
func tryParseDate(s string) (time.Time, error) {
	layouts := []string{
		time.RFC3339,
		"2006-01-02",
		"2006-01-02 15:04:05",
		"02-Jan-2006",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, errors.New("not a date")
}
