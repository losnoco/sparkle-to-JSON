package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
)

type Sparkle struct {
	XMLName xml.Name `xml:"rss"`
	Text    string   `xml:",chardata"`
	Sparkle string   `xml:"sparkle,attr"`
	Version string   `xml:"version,attr"`
	Channel struct {
		Text  string `xml:",chardata" json:"-"`
		Title string `xml:"title"`
		Item  []struct {
			Text                 string `xml:",chardata" json:"-"`
			Title                string `xml:"title"`
			PubDate              string `xml:"pubDate"`
			Version              string `xml:"version"`
			ShortVersionString   string `xml:"shortVersionString"`
			Description          string `xml:"description"`
			MinimumSystemVersion string `xml:"minimumSystemVersion"`
			Enclosure            struct {
				Text        string `xml:",chardata" json:"-"`
				URL         string `xml:"url,attr"`
				Length      string `xml:"length,attr"`
				Type        string `xml:"type,attr"`
				EdSignature string `xml:"edSignature,attr"`
			} `xml:"enclosure"`
			Deltas struct {
				Text      string `xml:",chardata" json:"-"`
				Enclosure []struct {
					Text               string `xml:",chardata" json:"-"`
					URL                string `xml:"url,attr"`
					DeltaFrom          string `xml:"deltaFrom,attr"`
					Length             string `xml:"length,attr"`
					Type               string `xml:"type,attr"`
					EdSignature        string `xml:"edSignature,attr"`
					Version            string `xml:"version,attr"`
					ShortVersionString string `xml:"shortVersionString,attr"`
				} `xml:"enclosure"`
			} `xml:"deltas" json:"-"`
			ReleaseNotesLink string `xml:"releaseNotesLink"`
		} `xml:"item"`
	} `xml:"channel"`
}

func main() {

	//if there is no argument, print usage
	if len(os.Args) < 2 {
		fmt.Println("Usage: sparkleToJSON <url>")
		os.Exit(1)
	}

	fmt.Println("Starting...")

	//get feed from file whose name is passed as argument
	var fileName string = os.Args[1]
	fmt.Println("Reading file: " + fileName)
	file, err := os.Open(fileName)

	var s Sparkle

	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	fmt.Println("Decoding XML...")
	if err := xml.NewDecoder(file).Decode(&s); err != nil {
		fmt.Println(err)
		return
	}

	items := s.Channel.Item

	fmt.Println("Sorting items...")
	//sort the items by short version string, converting to integers, descending
	sort.Slice(items, func(i, j int) bool {
		iInt, _ := strconv.Atoi(items[i].ShortVersionString)
		jInt, _ := strconv.Atoi(items[j].ShortVersionString)
		return iInt > jInt
	})

	//for each item, if Description is empty, get the release notes from ReleaseNotesLink
	fmt.Println("Getting release notes for items with empty descriptions...")
	for i := range items {
		if items[i].Description == "" && items[i].ReleaseNotesLink != "" {
			fmt.Printf(" - Getting release notes for version %s (%d of %d)... \n", items[i].Title, i+1, len(items))
			resp, err := http.Get(items[i].ReleaseNotesLink)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer resp.Body.Close()
			desc, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println(err)
				return
			}
			items[i].Description = string(desc)

		}
	}

	fmt.Println("Writing JSON...")
	//save to json file
	jsonFile, err := os.Create("mercury.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer jsonFile.Close()
	enc := json.NewEncoder(jsonFile)
	enc.SetIndent("", "  ")
	enc.Encode(items)

	fmt.Println("Done!")

}
