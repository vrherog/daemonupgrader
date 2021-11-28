package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

func RequestText(url, method, body string) (result string, statusCode uint16, err error) {
	var client = &http.Client{}
	var req *http.Request
	if body != `` {
		if req, err = http.NewRequest(method, url, bytes.NewBuffer([]byte(body))); err != nil {
			return
		}
	} else {
		if req, err = http.NewRequest(method, url, nil); err != nil {
			return
		}
	}
	var resp *http.Response
	if resp, err = client.Do(req); err != nil {
		return
	}
	statusCode = uint16(resp.StatusCode)
	var buffer []byte
	if buffer, err = ioutil.ReadAll(resp.Body); err != nil {
		return
	}
	defer resp.Body.Close()
	result = strings.TrimSpace(string(buffer))
	return
}

func RequestJson(url, method string, params []interface{}, reply interface{}) (statusCode uint16, err error) {
	var client = &http.Client{}
	var req *http.Request
	if params != nil && len(params) > 0 {
		var body []byte
		if body, err = json.Marshal(params); err != nil {
			return
		}
		if req, err = http.NewRequest(method, url, bytes.NewBuffer(body)); err != nil {
			return
		}
	} else {
		if req, err = http.NewRequest(method, url, nil); err != nil {
			return
		}
	}
	req.Header.Set(`Content-Type`, `application/json`)
	var resp *http.Response
	if resp, err = client.Do(req); err != nil {
		return
	}
	statusCode = uint16(resp.StatusCode)
	var buffer []byte
	if buffer, err = ioutil.ReadAll(resp.Body); err != nil {
		return
	}
	defer resp.Body.Close()
	if err = json.Unmarshal(buffer, &reply); err != nil {
		return
	}
	return
}

func DownloadFile(filename, url, method, body string) (statusCode uint16, err error) {
	var fp *os.File
	fp, err = os.Create(filename)
	if err != nil {
		return
	}
	defer fp.Close()
	var client = &http.Client{}
	var req *http.Request
	if body != `` {
		if req, err = http.NewRequest(method, url, bytes.NewBuffer([]byte(body))); err != nil {
			return
		}
	} else {
		if req, err = http.NewRequest(method, url, nil); err != nil {
			return
		}
	}
	var resp *http.Response
	if resp, err = client.Do(req); err != nil {
		return
	}
	statusCode = uint16(resp.StatusCode)
	if _, err = io.Copy(fp, resp.Body); err != nil {
		return
	}
	return
}
