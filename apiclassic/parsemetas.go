package apiclassic

import "github.com/PuerkitoBio/goquery"

type metaType uint8
const (
	metaUnknown = metaType(iota)
	metaProperty
	metaName
	metaItemProp
)

type metaTag struct {
	typ metaType
	name string
	content string
}

func enumMetas(s *goquery.Selection, next func(metaTag)bool) {
	// For each <meta>
	s.Find("meta").EachWithBreak(func(i int, s *goquery.Selection) bool {
		tag := metaTag{ metaUnknown, "", "" }
		for _, attr := range s.Nodes[0].Attr {
			switch attr.Key {
				case "property":
					tag.typ = metaProperty
					tag.name = attr.Val
				case "itemprop":
					tag.typ = metaItemProp
					tag.name = attr.Val
				case "name":
					tag.typ = metaName
					tag.name = attr.Val
				case "content":
					tag.content = attr.Val
			}

			if tag.typ == metaUnknown { continue }
			if len(tag.content) == 0 { continue }

			// Callback tag
			if !next(tag) {
				return false
			}
		}
		return true
	})
}