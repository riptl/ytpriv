package yt

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExtractVideoID(t *testing.T) {
	id := "JNhpwY5Zkzk"
	correct := []string{
		"https://www.youtube.com/watch?v=JNhpwY5Zkzk",
		"youtube.com/watch?v=JNhpwY5Zkzk",
		"https://youtube.com/watch?v=JNhpwY5Zkzk",
		"https://youtu.be/JNhpwY5Zkzk",
		"http://www.youtube.com/v/JNhpwY5Zkzk?version=3&amp;autohide=1",
		"https://www.youtube.com/embed/JNhpwY5Zkzk",
		"http://youtube.com/v/JNhpwY5Zkzk?version=3&amp;autohide=1",
		"https://youtube.com/embed/JNhpwY5Zkzk",
		"JNhpwY5Zkzk",
	}
	malformed := []string{
		"youtube.com/watch?v",
		"",
	}

	for _, test := range correct {
		res, err := ExtractVideoID(test)
		if err != nil {
			t.Errorf("When extracting \"%s\":\n\t%s", test, err)
		} else if res != id {
			t.Errorf("Expected: \"%s\", got: \"%s\" from \"%s\"", id, res, test)
		}
	}
	for _, test := range malformed {
		res, err := ExtractVideoID(test)
		if err == nil {
			t.Errorf("Extracted \"%s\" from malformed input \"%s\"", res, test)
		}
	}
}

func TestExtractChannelID(t *testing.T) {
	correct := []string{
		"https://www.youtube.com/channel/UCsLiV4WJfkTEHH0b9PmRklw",
		"http://www.youtube.com/channel/UCsLiV4WJfkTEHH0b9PmRklw/",
		"UCsLiV4WJfkTEHH0b9PmRklw",
	}
	malfored := []string{
		"https://www.youtubee.com/channel/UCsLiV4WJfkTEHH0b9PmRklw",
	}

	for _, test := range correct {
		id, err := ExtractChannelID(test)
		assert.Nilf(t, err, "Error parsing \"%s\": %s", test, err)
		assert.Equal(t, id, "UCsLiV4WJfkTEHH0b9PmRklw")
	}

	for _, test := range malfored {
		_, err := ExtractChannelID(test)
		assert.NotNil(t, err)
	}
}
