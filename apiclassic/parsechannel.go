package apiclassic

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/terorie/yt-mango/data"
	"github.com/valyala/fasthttp"
	"strconv"
)

func ParseChannel(c *data.Channel, res *fasthttp.Response) (err error) {
	if res.StatusCode() != 200 {
		return fmt.Errorf("HTTP status %d", res.StatusCode())
	}

	buf := res.Body()
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(buf))
	if err != nil { return }

	p := parseChannelInfo{c, doc}
	return p.parse()
}

type parseChannelInfo struct {
	c *data.Channel
	doc *goquery.Document
}

func (p *parseChannelInfo) parse() error {
	if err := p.parseMetas();
		err != nil { return err }
	return nil
}

func (p *parseChannelInfo) parseMetas() error {
	p.doc.Find("head").RemoveFiltered("#watch-container")
	enumMetas(p.doc.Find("head").Find("meta"), func(tag metaTag)bool {
		content := tag.content
		switch tag.typ {
		case metaProperty:
			switch tag.name {
			case "og:title":
				p.c.Name = content
			}
		case metaItemProp:
			switch tag.name {
			case "paid":
				if val, err := strconv.ParseBool(content);
					err == nil { p.c.Paid = val }
			}
		}
		return false
	})
	return nil
}

/*func (p *parseChannelInfo) parseAbout() error {
	p.doc.Find(".about-stats").Find(".about-stat").Each(func(_ int, s *goquery.Selection) {
		text := s.Text()
	})
	return nil
}*/
