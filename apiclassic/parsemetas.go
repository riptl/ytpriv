package apiclassic

import "github.com/PuerkitoBio/goquery"

type metaType uint8
const (
	metaUnknown = metaType(iota)
	metaProperty
	metaItemProp
)

type metaTag struct {
	typ metaType
	name string
	content string
}

func enumMetas(s *goquery.Selection, next func(metaTag)bool) {
	// For each <meta>
	s.EachWithBreak(func(i int, s *goquery.Selection) bool {
		tag := metaTag{ metaUnknown, "", "" }
		listAttrs: for _, attr := range s.Nodes[0].Attr {
			switch attr.Key {
				case "property":
					tag.typ = metaProperty
					tag.name = attr.Val
					break listAttrs
				case "itemprop":
					tag.typ = metaItemProp
					tag.name = attr.Val
					break listAttrs
				case "content":
					tag.content = attr.Val
					break listAttrs
			}

			if tag.typ == metaUnknown { continue }
			if len(tag.content) == 0 { continue }

			// Callback tag
			if !next(tag) {
				return true
			}
		}
		return false
	})
}