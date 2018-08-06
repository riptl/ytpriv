package api

var MarkdownTextEscape EscapeMap
var MarkdownLinkEscape EscapeMap

func init() {
	registerMap := func(eMap EscapeMap, escaped string) {
		for _, c := range escaped {
			eMap.Set(uint(c), true)
		}
	}

	registerMap(MarkdownTextEscape, "\\!\"#$%&()*+/;<=>?@[]^_`{|}~-")
	registerMap(MarkdownLinkEscape, "\\!\"#$%&'()*+,;<=>?@[]^_`{|}~-")
}

