package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/go-resty/resty/v2"
)

var version string

type config []account

type account struct {
	Email    string
	Password string
	Videos   []video
}

type video struct {
	Title              string
	Description        string
	PublishNow         bool
	ContentSensitivity string
	Cover              string
	Video              string
}

type uploadData struct {
	uploadCode string
	cid        string
	cdid       string
}

var exampleconfig = []byte(`[
	{
		"email": "Email Here",
		"password": "Password Here",
		"videos": [
			{
				"title": "Title Here",
				"description": "Description Here",
				"publishNow": true,
				"contentSensitivity": "normal",
				"cover": "Full Path to Cover Image Here",
				"video": "Full Path to Video File Here"
			}
		]
	}
]`)

func main() {
	os.Exit(mainExitCode())
}

func mainExitCode() int {
	var v1 bool
	flag.BoolVar(&v1, "version", false, "Print version number and exit")
	flag.BoolVar(&v1, "v", false, "Print version number and exit")
	flag.Parse()
	if v1 {
		fmt.Println(version)
		return 0
	}

	var err error
	file, err := os.Open("config.json")
	if err != nil {
		if err.Error() == "open config.json: The system cannot find the file specified." {
			fmt.Println("No config file found, creating example config")
			err := ioutil.WriteFile("config.json", exampleconfig, 0644)
			if err != nil {
				fmt.Println("Could not create example config")
				fmt.Println(err)
				return 1
			}
			return 0
		}
		fmt.Println(err)
		return 1
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		// TODO proper error message / when does this happen?
		fmt.Println(err)
		return 1
	}
	var config config
	err = json.Unmarshal(data, &config)
	if err != nil {
		fmt.Println("Could not parse config.json, it may be corrupted")
		fmt.Println(err)
		return 1
	}

	for i := 0; i < len(config); i++ {
		handleAccount(config[i])
	}
	return 0
}

func handleAccount(account account) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error during account login:")
			fmt.Println(r)
		}
	}()

	fmt.Printf("Logging into account \"%v\"\n", account.Email)
	// log into account
	client := resty.New()
	client.SetRedirectPolicy(resty.NoRedirectPolicy())

	// get cookies and csrfmiddlewaretoken
	resp, err := client.R().Get("https://www.bitchute.com/")
	if err != nil {
		panic(err)
	}
	csrfRegex := regexp.MustCompile("name='csrfmiddlewaretoken' value='(.+?)'")
	csrfMatch := csrfRegex.FindStringSubmatch(resp.String())
	if csrfMatch == nil {
		panic(errors.New("Homepage: couldn't find csrfmiddlewaretoken"))
	}
	csrfToken := csrfMatch[1]

	resp, err = client.R().
		SetHeader("Referer", "https://www.bitchute.com/").
		SetFormData(map[string]string{
			"csrfmiddlewaretoken": csrfToken,
			"username":            account.Email,
			"password":            account.Password,
			"one_time_code":       "",
		}).
		Post("https://www.bitchute.com/accounts/login/")
	if err != nil {
		panic(err)
	}
	expectStatus(resp, 200)

	for j := 0; j < len(account.Videos); j++ {
		handleVideo(client, csrfToken, account.Videos[j])
	}
}

func handleVideo(client *resty.Client, csrfToken string, video video) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error during video upload:")
			fmt.Println(r)
		}
	}()

	fmt.Printf("Uploading video \"%v\"\n", video.Title)
	// start upload
	resp, err := client.R().SetHeader("Referer", "https://www.bitchute.com/").Get("https://www.bitchute.com/myupload/")
	if err != nil {
		if match, _ := regexp.MatchString(`auto redirect is disabled(?:$|\n)`, err.Error()); !match {
			panic(err)
		}
	}
	expectStatus(resp, 302)
	uploadLink := resp.Header().Get("location")

	u, _ := url.Parse(uploadLink)
	m, _ := url.ParseQuery(u.RawQuery)
	uploadData := parseUploadData(m)

	baseUploadLink := uploadLink[:strings.Index(uploadLink, "?")]
	metaUploadLink := strings.Replace(baseUploadLink, "/upload/", "/uploadmeta/", 1)
	finishUploadLink := strings.Replace(baseUploadLink, "/upload/", "/finish_upload/", 1)

	// get cookies and csrfmiddlewaretoken
	resp, err = client.R().Get(uploadLink)
	if err != nil {
		panic(err)
	}
	csrfRegex := regexp.MustCompile("name='csrfmiddlewaretoken' value='(.+?)'")
	csrfMatch := csrfRegex.FindStringSubmatch(resp.String())
	if csrfMatch == nil {
		panic(errors.New("Uploader: couldn't find csrfmiddlewaretoken"))
	}
	csrfToken = csrfMatch[1]

	fmt.Println("Uploading cover image")
	// upload cover image
	file, err := os.Open(video.Cover)
	if err != nil {
		panic(err)
	}
	contentType := resty.DetectContentType(file)
	file.Seek(0, 0)
	resp, err = client.R().
		SetHeader("Referer", uploadLink).
		SetFormData(map[string]string{
			"csrfmiddlewaretoken": csrfToken,
			"upload_code":         uploadData.uploadCode,
			"upload_type":         "video",
		}).
		SetMultipartField("file", file.Name(), contentType, file).
		Post(baseUploadLink)
	if err != nil {
		panic(err)
	}
	expectStatus(resp, 200)

	fmt.Println("Uploading video file")
	// upload video file
	file, err = os.Open(video.Video)
	if err != nil {
		panic(err)
	}
	contentType = resty.DetectContentType(file)
	file.Seek(0, 0)
	resp, err = client.R().
		SetHeader("Referer", uploadLink).
		SetFormData(map[string]string{
			"csrfmiddlewaretoken": csrfToken,
			"upload_code":         uploadData.uploadCode,
			"upload_type":         "video",
		}).
		SetMultipartField("file", file.Name(), contentType, file).
		Post(baseUploadLink)
	if err != nil {
		panic(err)
	}
	expectStatus(resp, 200)

	fmt.Println("Setting video title and description")
	// set upload info
	resp, err = client.R().
		SetHeader("Referer", uploadLink).
		SetFormData(map[string]string{
			"csrfmiddlewaretoken": csrfToken,
			"upload_title":        video.Title,
			"upload_description":  video.Description,
			"upload_code":         uploadData.uploadCode,
		}).
		Post(metaUploadLink)
	if err != nil {
		panic(err)
	}

	fmt.Println("Finishing upload")
	// finish upload
	resp, err = client.R().
		SetHeader("Referer", uploadLink).
		SetFormData(map[string]string{
			"csrfmiddlewaretoken": csrfToken,
			"upload_code":         uploadData.uploadCode,
			"cid":                 uploadData.cid,
			"cdid":                uploadData.cdid,
			"sensitivity":         parseContentSensitivity(video.ContentSensitivity),
			"publish_now":         parsePublishNow(video.PublishNow),
		}).
		Post(finishUploadLink)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Video \"%v\" uploaded\nVideo link: https://www.bitchute.com/video/%v/\n", video.Title, uploadData.uploadCode)
}

/* utility functions */

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
