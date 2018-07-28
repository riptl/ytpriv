package browseajax

func GetPage(channelID string, page uint) error {
	root, err := GrabPage(channelID, page)
	if err != nil { return err }
	err = ParsePage(root)
	if err != nil { return err }
	return nil
}
