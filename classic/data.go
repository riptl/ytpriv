package classic

type XMLSubTrackList struct {
	Tracks []struct {
		LangCode string `xml:"lang_code,attr"`
		Lang     string `xml:"lang_translated,attr"`
	} `xml:"track"`
}
