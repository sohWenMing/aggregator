package rss_parsing

import (
	"encoding/xml"
	"html"

	"github.com/microcosm-cc/bluemonday"
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

	htmlParser := bluemonday.StrictPolicy()

	returnedFeed := RSSFeed{}
	unMarshalErr := xml.Unmarshal(buf, &returnedFeed)
	if unMarshalErr != nil {
		return RSSFeed{}, unMarshalErr
	}
	for i, item := range returnedFeed.Channel.RSSItems {
		returnedFeed.Channel.RSSItems[i].Title = html.UnescapeString(item.Title)
		returnedFeed.Channel.RSSItems[i].Description = htmlParser.Sanitize(html.UnescapeString(item.Description))
	}
	returnedFeed.Channel.Title = html.UnescapeString(returnedFeed.Channel.Title)
	returnedFeed.Channel.Description = html.UnescapeString(returnedFeed.Channel.Description)
	return returnedFeed, nil
}
