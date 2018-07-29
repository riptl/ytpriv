package browseajax

func GetPage(channelID string, page uint) ([]string, error) {
	root, err := GrabPage(channelID, page)
	if err != nil { return nil, err }
	urls, err := ParsePage(root)
	if err != nil { return nil, err }
	return urls, nil
}
