# WIP: youtube-mango 💾

> YouTube metadata extractor inspired by [`youtube-ma` by _CorentinB_][1]

##### Build

Install and compile the Go project with `go get github.com/terorie/youtube-mango`!

If you don't have a Go toolchain, grab an executable from the Releases tab

##### Project structure

- _/common_: commonly used HTTP code
- _/data_: Data structures
- _/classic_: Extractor calling the HTML `/watch` API
- _/watchapi_: Extractor calling the JSON `/watch` API

 [1]: https://github.com/CorentinB/youtube-ma
