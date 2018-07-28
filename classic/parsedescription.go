package classic

import (
	"errors"
	"golang.org/x/net/html"
	"bytes"
	"github.com/terorie/youtube-mango/common"
	"strings"
)

const descriptionSelector = "#eow-description"

func (p *parseInfo) parseDescription() error {
	// Find description root
	descNode := p.doc.Find(descriptionSelector).First()
	if len(descNode.Nodes) == 0 { return errors.New("could not find description") }

	// Markdown text
	var buffer bytes.Buffer

	// Enumerate nodes
	for c := descNode.Nodes[0].FirstChild; c != nil; c = c.NextSibling {
		switch c.Type {
		case html.TextNode:
			// FIXME: "&amp;lt;" gets parsed to => "<"
			// Write text to buffer, escaping markdown
			err := common.MarkdownTextEscape.ToBuffer(c.Data, &buffer)
			if err != nil { return err }
		case html.ElementNode:
			switch c.Data {
			// Newline
			case "br":
				err := buffer.WriteByte(0x0a)
				if err != nil { return err }
			// Link
			case "a":
				err := parseLink(c, &buffer)
				if err != nil { return err }
			}
		}
	}

	// Save description
	p.v.Description = buffer.String()
	println(p.v.Description)

	return nil
}

func parseLink(c *html.Node, dest *bytes.Buffer) error {
	// Find text
	if c.FirstChild == nil { return nil } // Empty link
	if c.FirstChild.Type != html.TextNode {
		return errors.New("unexpected non-text node")
	}
	text := c.FirstChild.Data

	// Find href
	for _, attr := range c.Attr {
		if attr.Key == "href" {
			switch {
			// hashtag
			case strings.HasPrefix(attr.Val, "/results"):
				dest.WriteString(text)

			// real link
			case strings.HasPrefix(attr.Val, "/redirect"):
				/*
				Not needed:
				// Decode link from href
				link, err := decodeLink(attr.Val)
				if err != nil { return err }
				// Escape to markdown
				link, err = common.MarkdownLinkEscape.ToString(link)
				if err != nil { return err }
				// Write to buffer
				dest.WriteString(fmt.Sprintf("[%s](%s)\n", text, link))
				*/
				dest.WriteString(text)

			default:
				return errors.New("unknown link")
			}
			break
		}
	}
	return nil
}

/* Not needed

func decodeLink(href string) (string, error) {
	url, err := url2.Parse(href)
	if err != nil { return "", err }

	query := url.Query()
	link := query.Get("q")
	if link == "" { return "", errors.New("empty link") }

	return link, nil
}

 */
