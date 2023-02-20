package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gocolly/colly"
)

const StartURL = "https://www.tamiya.com/japan/event/index.html?genre_item=event_category%2cevent_type%2cevent_pref&cmdarticlesearch=1&absolutepage=1&field_sort=d&sortkey=sort_posd"

func main() {
	c := colly.NewCollector()
	rss := &RSS{
		Version:     "2.0",
		Title:       "タミヤ イベント",
		Link:        StartURL,
		Description: "タミヤ イベント 掲載順",
	}

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.OnHTML(".event_calendar_ tr", func(e *colly.HTMLElement) {
		url, found := e.DOM.Find("a").Attr("href")
		if !found {
			return
		}
		date := strings.TrimSpace(e.DOM.Find("th").Text())
		title := strings.TrimSpace(e.DOM.Find(".ttl_").Text())
		pref := strings.TrimSpace(e.DOM.Find(".pref_ span").Text())

		description := []string{}

		e.ForEach(".genre_filter_category span", func(_ int, e *colly.HTMLElement) {
			description = append(description, strings.TrimSpace(e.Text))
		})

		e.ForEach(".point_ span", func(_ int, e *colly.HTMLElement) {
			description = append(description, strings.TrimSpace(e.Text))
		})

		rss.AppendItem(Item{
			Title:       strings.Join([]string{date, pref, title}, " "),
			Link:        url,
			Description: strings.Join(description, ", "),
		})
	})

	c.OnHTML("a[rel=\"next\"]", func(e *colly.HTMLElement) {
		href := e.Attr("href")
		c.Visit(e.Request.AbsoluteURL(href))
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	c.OnScraped(func(r *colly.Response) {
		finish(rss)
	})

	c.Visit(StartURL)
}

func finish(rss *RSS) {
	data, _ := xml.MarshalIndent(rss, "", "  ")

	fmt.Println(string(data))

	sess := session.Must(session.NewSession())
	uploader := s3manager.NewUploader(sess)

	bucket := os.Getenv("S3_BUCKET")

	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String("tamiya-events.xml"),
		Body:        aws.ReadSeekCloser(bytes.NewReader(data)),
		ACL:         aws.String("public-read"),
		ContentType: aws.String("application/rss+xml"),
	})

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result.UploadID)
}
