package browseajax

import (
	"github.com/valyala/fastjson"
	"errors"
)

var missingData = errors.New("missing data")

func ParsePage(rootObj *fastjson.Value) error {
	// Root as array
	root, err := rootObj.Array()
	if err != nil { return err }

	// Find response container
	var container *fastjson.Value
	for _, item := range root {
		if item.Exists("response") {
			container = item
			break
		}
	}
	if container == nil { return missingData }

	// Get error obj

	// Get items from grid
	itemsObj := container.Get(
		"response",
		"continuationContents",
		"gridContinuation",
		"items",
	)
	if itemsObj == nil { return missingData }

	// Items as array
	items, err := itemsObj.Array()
	if err != nil { return err }

	// Enumerate
	for _, item := range items {
		// Find URL
		urlObj := item.Get(
			"gridVideoRenderer",
			"navigationEndpoint",
			"commandMetadata",
			"webCommandMetadata",
			"url",
		)
		if urlObj == nil { return missingData }

		// URL as string
		urlBytes, err := urlObj.StringBytes()
		if err != nil { return err }
		url := string(urlBytes)

		println(url)
	}
	return nil
}
