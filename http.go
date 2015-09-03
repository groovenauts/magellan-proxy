package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var transport = http.Transport{
	DisableKeepAlives: true,
}
var Port int = 80
var BaseUrl string

func InitHttpTransport(port, num int) {
	transport.DisableKeepAlives = true
	transport.MaxIdleConnsPerHost = num
	Port = port
	BaseUrl = "http://127.0.0.1:" + strconv.Itoa(Port)
}

func ProcessHttpRequest(req *Request) (*Response, error) {
	url := BaseUrl + req.Env.PathInfo
	url += req.Env.QueryString
	httpReq, err := http.NewRequest(req.Env.Method, url, bytes.NewReader(req.Body))
	if err != nil {
		log.Printf("cannot create HTTP Request: %s")
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
		log.Printf("HTTP request failed: %s", err.Error())
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
		log.Printf("HTTP Response body read fail: %s", err.Error())
		return nil, err
	}
	res.Body = b
	res.BodyEncoding = "plain"
	return res, nil
}

// vim:set noexpandtab ts=2:
