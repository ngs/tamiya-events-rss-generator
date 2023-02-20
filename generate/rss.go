package main

import "encoding/xml"

type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
}

type RSS struct {
	XMLName     xml.Name `xml:"rss"`
	Version     string   `xml:"version,attr"`
	Description string   `xml:"channel>description"`
	Link        string   `xml:"channel>link"`
	Title       string   `xml:"channel>title"`
	Items       []Item   `xml:"channel>item"`
}

func (rss *RSS) AppendItem(item Item) {
	rss.Items = append(rss.Items, item)
}
