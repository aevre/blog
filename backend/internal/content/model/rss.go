package contentmodel

import "encoding/xml"

type RSS struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Channel RssChannel `xml:"channel"`

	XMLNs   string `xml:"xmlns:atom,attr"`
	Content string `xml:"xmlns:content,attr"`
	DC      string `xml:"xmlns:dc,attr"`
	FH      string `xml:"xmlns:fh,attr,omitempty"`
}

type RssChannel struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	LastBuild   string `xml:"lastBuildDate"`

	Items     []RssItem   `xml:"item"`
	AtomLinks []AtomLink  `xml:"atom:link"`
	Complete  *FhComplete `xml:"fh:complete,omitempty"`
}

type AtomLink struct {
	Href  string `xml:"href,attr"`
	Rel   string `xml:"rel,attr"`
	Type  string `xml:"type,attr"`
	Title string `xml:"title,attr,omitempty"`
}

type FhComplete struct{}

type RssItemDescription struct {
	Value string `xml:",cdata"`
}

type RssItemContent struct {
	Value string `xml:",cdata"`
}

type RssCategory struct {
	Domain string `xml:"domain,attr,omitempty"`
	Value  string `xml:",chardata"`
}

type RssGUID struct {
	Value       string `xml:",chardata"`
	IsPermaLink bool   `xml:"isPermaLink,attr"`
}

type RssItem struct {
	Title       string             `xml:"title"`
	Link        string             `xml:"link"`
	GUID        RssGUID            `xml:"guid"`
	PubDate     string             `xml:"pubDate"`
	Description RssItemDescription `xml:"description"`

	Content    *RssItemContent `xml:"content:encoded,omitempty"`
	Author     string          `xml:"dc:creator,omitempty"`
	Categories []RssCategory   `xml:"category,omitempty"`
}
