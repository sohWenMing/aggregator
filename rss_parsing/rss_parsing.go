package rss_parsing

import (
	"encoding/xml"
	"html"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		RSSItems    []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func ParseRSS(buf []byte) (rssFeed RSSFeed, err error) {

	returnedFeed := RSSFeed{}
	unMarshalErr := xml.Unmarshal(buf, &returnedFeed)
	if unMarshalErr != nil {
		return RSSFeed{}, unMarshalErr
	}
	for _, item := range returnedFeed.Channel.RSSItems {
		item.Title = html.UnescapeString(item.Title)
		item.Description = html.UnescapeString(item.Description)
	}
	returnedFeed.Channel.Title = html.UnescapeString(returnedFeed.Channel.Title)
	returnedFeed.Channel.Description = html.UnescapeString(returnedFeed.Channel.Description)
	return returnedFeed, nil
}
