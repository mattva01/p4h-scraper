package main

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/headzoo/surf"
	"net/url"
	"os"
	"runtime"
	"strings"
)

type Entry struct {
	Pid          string `json:"pid"`
	Title        string `json:"title"`
	PictureUrl   string `json:"picture_url"`
	Organization string `json:"org"`
	Summary      string `json:"summary"`
	Category     string `json:"category"`
	Type         string `json:"type"`
	Location     string `json:"location"`
	Contact      string `json:"contact"`
}

// Individual project pages
func parseDetail(e *Entry, c chan int) {
	browser := surf.NewBrowser()
	browser.Open("https://p4h.skild.com/skild2/p4h/viewEntryDetail.action?pid=" + e.Pid)
	form := browser.Find("#form-builder")
	e.Title = strings.TrimSpace(form.Find("div:nth-child(3) > span").Last().Text())
	e.Organization = strings.TrimSpace(form.Find("div:nth-child(6) > span").Last().Text())
	e.Summary = strings.TrimSpace(form.Find("div:nth-child(8) > span").Last().Text())
	e.PictureUrl, _ = form.Find("div:nth-child(10) img").Attr("src")
	e.Category = strings.TrimSpace(form.Find("div:nth-child(12) input[checked='checked'] ~ label").Text())
	e.Type = strings.TrimSpace(form.Find("div:nth-child(14) input[checked='checked'] ~ label").Text())
	e.Location = strings.TrimSpace(form.Find("div:nth-child(16) > span").Last().Text())
	e.Contact = strings.TrimSpace(form.Find("div:nth-child(17) > span").Last().Text())
	c <- 1
}

func main() {
	//Effective multicore
	runtime.GOMAXPROCS(runtime.NumCPU())

	var entrylist []*Entry

	browser := surf.NewBrowser()

	// I'm still not sure why this is required... gotta love chunked encoding
	browser.Open("https://p4h.skild.com/skild2/p4h/viewEntryVoting.action")
	browser.Open("https://p4h.skild.com/skild2/p4h/publicVotingGetEntryList.action?sortBy=title&filteryBy=0&viewType=list")

	length := browser.Find("div.entry").Length()

	//Get pids
	browser.Find("div.entry").Each(func(_ int, s *goquery.Selection) {
		entry := new(Entry)
		entrylist = append(entrylist, entry)

		// Get pid
		pidurl, _ := s.Find("div.post-content > h4 > a").Attr("href")
		u, _ := url.Parse(pidurl)
		pid := u.Query().Get("pid")
		entry.Pid = pid
	})

	// Retrieve all the detail pages at once
	c := make(chan int)
	for i := 0; i < length; i++ {
		go parseDetail(entrylist[i], c)
	}

	// Wait for them all to complete
	for i := 0; i < length; i++ {
		<-c
		fmt.Printf("Processed %d of %d\n", i+1, length)
	}

	// Write out the json file
	out, _ := json.MarshalIndent(entrylist, "", "    ")
	jsonFile, err := os.Create("./p4h.json")
	if err != nil {
		fmt.Println(err)
	}
	jsonFile.Write(out)
	jsonFile.Close()

}
