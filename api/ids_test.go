package api

import "testing"

func TestGetVideoID(t *testing.T) {
	id := "JNhpwY5Zkzk"
	correct := []string {
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
	malformed := []string {
		"youtube.com/watch?v",
		"",
	}

	for _, test := range correct {
		res, err := GetVideoID(test)
		if err != nil {
			t.Errorf("When extracting \"%s\":\n\t%s", test, err)
		} else if res != id {
			t.Errorf("Expected: \"%s\", got: \"%s\" from \"%s\"", id, res, test)
		}
	}
	for _, test := range malformed {
		res, err := GetVideoID(test)
		if err == nil {
			t.Errorf("Extracted \"%s\" from malformed input \"%s\"", res, test)
		}
	}
}
