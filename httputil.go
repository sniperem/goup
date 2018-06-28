package goup

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func NewHttpRequest(client *http.Client, method string, url string, postData string, headers map[string]string) ([]byte, error) {
	req, _ := http.NewRequest(method, url, strings.NewReader(postData))
	if headers != nil {
		for k, v := range headers {
			req.Header.Add(k, v)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	bodyData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HttpStatusCode: %d, reply: %s", resp.StatusCode, string(bodyData))
	}

	return bodyData, nil
}

func HttpGet(client *http.Client, url string) (map[string]interface{}, error) {
	respData, err := NewHttpRequest(client, "GET", url, "", nil)
	if err != nil {
		return nil, err
	}

	var bodyDataMap map[string]interface{}
	//fmt.Printf("\n%s\n", respData);
	err = json.Unmarshal(respData, &bodyDataMap)
	if err != nil {
		log.Println(string(respData))
		return nil, err
	}
	return bodyDataMap, nil
}
