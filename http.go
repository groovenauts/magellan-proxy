package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"
)

var transport = http.Transport{
	DisableKeepAlives: true,
}

func InitHttpTransport() {
	transport.DisableKeepAlives = true
	transport.MaxIdleConnsPerHost = 0
}

func ProcessHttpRequest(req *Request) (*Response, error) {
	url := "http://127.0.0.1" + req.Env.PathInfo
	url += req.Env.QueryString
	httpReq, err := http.NewRequest(req.Env.Method, url, bytes.NewReader(req.Body))
	if err != nil {
		println("magellan-proxy: cannot create HTTP Request:" + err.Error())
		return nil, err
	}
	for h, v := range req.Headers {
		httpReq.Header.Add(h, v)
	}
	httpReq.Header.Add("x-forwarded-host", req.Headers["host"])
	httpReq.Header.Add("x-forwarded-server", req.Env.ServerName)
	httpReq.Host = req.Headers["host"]
	tr := &transport
	httpRes, err := tr.RoundTrip(httpReq)
	if err != nil {
		println("magellan-proxy: HTTP request failed.", err.Error())
		return nil, err
	}
	defer httpRes.Body.Close()
	res := new(Response)
	res.Status = httpRes.Status
	res.Headers = make(map[string]string)
	for k, v := range httpRes.Header {
		res.Headers[k] = strings.Join(v, "\n")
	}
	b, err := ioutil.ReadAll(httpRes.Body)
	if err != nil {
		println("magellan-proxy: HTTP Response body read fail: ", err.Error())
		return nil, err
	}
	res.Body = b
	res.BodyEncoding = "plain"
	return res, nil
}

// vim:set noexpandtab ts=2:
