package main

import (
	"fmt"

	"github.com/go-resty/resty/v2"
)

func expectStatus(resp *resty.Response, status int) {
	if resp.StatusCode() != status {
		panic(fmt.Errorf("%v response when logging in\nBody: %v", resp.Status(), resp.String()))
	}
}

func parseContentSensitivity(contentSensitivity string) string {
	switch contentSensitivity {
	case "normal":
		return "10"
	case "nsfw":
		return "40"
	case "nsfl":
		return "70"
	default:
		fmt.Printf("Invalid content sensitivity \"%v\", defaulting to \"nsfl\" to be safe\n", contentSensitivity)
		return "70"
	}
}

func parsePublishNow(publishNow bool) string {
	if publishNow {
		return "true"
	}
	return "false"
}

func parseUploadData(m map[string][]string) uploadData {
	var data uploadData
	data.uploadCode = m["upload_code"][0]
	data.cid = m["cid"][0]
	data.cdid = m["cdid"][0]
	return data
}
