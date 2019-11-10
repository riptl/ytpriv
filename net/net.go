package net

import "github.com/valyala/fasthttp"

var MaxWorkers uint

var Client = fasthttp.Client{
	Name: "yt-mango/0.1",
	DisableHeaderNamesNormalizing: true,
	MaxConnsPerHost: 50,
}
