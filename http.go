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
var PublishPath string

func InitHttpTransport(port, num int, publishPath string) {
	transport.DisableKeepAlives = true
	transport.MaxIdleConnsPerHost = num
	Port = port
	BaseUrl = "http://127.0.0.1:" + strconv.Itoa(Port)
	PublishPath = publishPath
}

func CreateHttpRequest(req *Request) (*http.Request, error) {
	if req.Env.Method == "PUBLISH" {
		url := BaseUrl + PublishPath
		url += "?topic=" + req.Env.PathInfo
		httpReq, err := http.NewRequest("POST", url, bytes.NewReader(req.Body))
		if err != nil {
			return nil, err
		}
		httpReq.ContentLength = int64(len(req.Body))
		httpReq.Header.Add("x-magellan-mqtt", "1")
		httpReq.Header.Add("content-type", "application/octet-stream")
		return httpReq, nil
	} else {
		url := BaseUrl + req.Env.PathInfo
		url += req.Env.QueryString
		httpReq, err := http.NewRequest(req.Env.Method, url, bytes.NewReader(req.Body))
		if err != nil {
			return nil, err
		}
		for h, v := range req.Headers {
			httpReq.Header.Add(h, v)
		}
		httpReq.Header.Add("x-forwarded-host", req.Headers["host"])
		httpReq.Header.Add("x-forwarded-server", req.Env.ServerName)
		httpReq.Host = req.Headers["host"]
		return httpReq, nil
	}
}

func ProcessHttpRequest(req *Request) (*Response, error) {
	httpReq, err := CreateHttpRequest(req)
	if err != nil {
		log.Printf("cannot create HTTP Request: %s")
		return nil, err
	}
	tr := &transport
	httpRes, err := tr.RoundTrip(httpReq)
	if err != nil {
		log.Printf("HTTP request failed: %s", err.Error())
		return nil, err
	}
	defer httpRes.Body.Close()
	if req.Env.Method == "PUBLISH" {
		_, err := ioutil.ReadAll(httpRes.Body)
		if err != nil {
			log.Printf("HTTP Response body read fail: %s", err.Error())
			return nil, err
		}
		return nil, nil
	} else {
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
}

// vim:set noexpandtab ts=2:
