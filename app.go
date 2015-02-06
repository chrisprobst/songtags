package songtags

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
)

func chromaprint(path string, raw bool) (duration, fingerprint string) {
	// Setup command
	var cmd *exec.Cmd
	if raw {
		cmd = exec.Command("fpcalc", "-raw", path)
	} else {
		cmd = exec.Command("fpcalc", path)
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

func lookupSong(duration, fingerprint string) string {
	url := "http://api.acoustid.org/v2/lookup?client=8XaBELgH&meta=recordings+releasegroups+compress&duration=" + duration + "&fingerprint=" + fingerprint
	response, err := http.Get(url)
	var finalRes string
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

		var buffer bytes.Buffer
		if lookup.Status == "ok" {
			for _, res := range lookup.Results {
				for _, rec := range res.Recordings {
					fmt.Fprintln(&buffer, "Song title:", rec.Title)
					fmt.Fprintln(&buffer, "Duration:", rec.Duration)
					fmt.Fprintln(&buffer, "Release groups:")
					for _, rg := range rec.Releasegroups {
						for _, art := range rg.Artists {
							fmt.Fprintln(&buffer, "    Artist:", art.Name)
						}

						fmt.Fprintln(&buffer, "        Type:", rg.Type)
						fmt.Fprintln(&buffer, "        Title:", rg.Title)
						fmt.Fprintln(&buffer)
					}
				}
			}
		}
		finalRes = string(buffer.Bytes())
	}
	return finalRes
}

func ForFile(fp string) string {
	return lookupSong(chromaprint(fp, false))
}
