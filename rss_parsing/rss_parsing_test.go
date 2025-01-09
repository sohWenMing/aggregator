package rss_parsing

import (
	"io"
	"os"
	"strings"
	"testing"

	testutils "github.com/sohWenMing/aggregator/test_utils"
)

func TestParseRSS(t *testing.T) {
	buf := getXMLBuf(t)
	rssFeed, err := ParseRSS(buf)
	testutils.AssertNoErr(err, t)
	testutils.AssertStrings(rssFeed.Channel.Title, "Lane's Blog", t)
	testutils.AssertStrings(rssFeed.Channel.Description, "Recent content on Lane's Blog", t)
	for _, item := range rssFeed.Channel.RSSItems {
		if strings.Contains(item.Description, "&amp") {
			t.Errorf("unescaped description: %s\n", item.Description)
		}
		if strings.Contains(item.Title, "&amp") {
			t.Errorf("unescaped description: %s\n", item.Title)
		}
	}
}

func getXMLBuf(t *testing.T) (buf []byte) {
	testFile, err := os.Open("testfile.xml")
	testutils.AssertNoErr(err, t)
	defer testFile.Close()

	read, err := io.ReadAll(testFile)
	testutils.AssertNoErr(err, t)
	return read
}
