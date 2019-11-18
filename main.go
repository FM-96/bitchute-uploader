package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
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
		account := config[i]
		fmt.Printf("Logging into account \"%v\"\n", account.Email)
		// TODO log into account
		for j := 0; j < len(account.Videos); j++ {
			video := account.Videos[j]
			fmt.Printf("Uploading video \"%v\"\n", video.Title)
			// TODO start upload
			fmt.Println("Uploading cover image")
			// TODO upload cover image
			fmt.Println("Uploading video file")
			// TODO upload video file
			fmt.Println("Setting video title and description")
			// TODO set upload info
			fmt.Println("Finishing upload")
			// TODO finish upload
		}
	}
	return 0
}
