package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func chromaprint(path string, raw bool) (duration, fingerprint string) {
	// Setup command
	var cmd *exec.Cmd
	if raw {
		cmd = exec.Command("fpcalc", "-raw", os.Args[1])
	} else {
		cmd = exec.Command("fpcalc", os.Args[1])
	}

	// Generate fingerprint
	fp, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}

	// Extract duration
	pattern := "DURATION=[0-9]*"
	re := regexp.MustCompile(pattern)
	duration = strings.Split(re.FindString(string(fp)), "=")[1]

	// Extract fingerprint
	pattern = "FINGERPRINT=.*"
	re = regexp.MustCompile(pattern)
	fingerprint = strings.Split(re.FindString(string(fp)), "=")[1]

	return
}

// Used for result fetching

type Lookup struct {
	Status  string
	Results []Results
}

type Results struct {
	Recordings []Recording
	Score      float64
	Id         string
}

type Recording struct {
	Artists       []Artist
	Duration      float64
	Releasegroups []Releasegroup
	Title         string
	Id            string
}

type Releasegroup struct {
	Artists        []Artist
	Secondarytypes []string
	Type           string
	Id             string
	Title          string
}

type Artist struct {
	Id   string
	Name string
}

func lookupSong(duration, fingerprint string) {
	url := "http://api.acoustid.org/v2/lookup?client=8XaBELgH&meta=recordings+releasegroups+compress&duration=" + duration + "&fingerprint=" + fingerprint
	response, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	} else {
		// Read web response
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Fatal(err)
		}

		// Decode json
		var lookup Lookup
		if err := json.Unmarshal(contents, &lookup); err != nil {
			log.Fatal(err)
		}

		if lookup.Status == "ok" {
			for _, res := range lookup.Results {
				for _, rec := range res.Recordings {
					fmt.Println("Song title:", rec.Title)
					fmt.Println("Duration:", rec.Duration)
					fmt.Println("Release groups:")
					for _, rg := range rec.Releasegroups {
						for _, art := range rg.Artists {
							fmt.Println("    Artist:", art.Name)
						}

						fmt.Println("        Type:", rg.Type)
						fmt.Println("        Title:", rg.Title)
						fmt.Println()
					}
				}
			}
		}
	}
}

func main() {
	lookupSong(chromaprint(os.Args[1], false))
}
