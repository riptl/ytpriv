package net

import "github.com/valyala/fasthttp"

var MaxWorkers uint

var Client = fasthttp.Client{
	Name: "ytwrk/v0.4",
	DisableHeaderNamesNormalizing: true,
	MaxConnsPerHost: 50,
}
