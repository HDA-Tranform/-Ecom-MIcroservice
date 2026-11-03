package handler

import (
	"encoding/xml"
	"io"
)

type DataList struct {
	XMLName xml.Name `xml:"DataList"`
	Items   []Data   `xml:"Data"`
}

type Data struct {
	Row    string
	Fields map[string]string
}

func (d *Data) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	d.Fields = make(map[string]string)
	for _, attr := range start.Attr {
		d.Fields[attr.Name.Local] = attr.Value
		if attr.Name.Local == "row" {
			d.Row = attr.Value
		}
	}
	// skip end element
	for {
		t, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if end, ok := t.(xml.EndElement); ok && end.Name == start.Name {
			break
		}
	}
	return nil
}
