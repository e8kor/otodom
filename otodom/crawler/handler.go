package function

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/url"

	"github.com/gocolly/colly/v2"

	otodom "github.com/e8kor/crawler/otodom/commons"

	handler "github.com/openfaas/templates-sdk/go-http"
)

func Handle(r handler.Request) (response handler.Response, err error) {
	var (
		entries      []otodom.Entry
		httpResponse *http.Response
	)
	query, err := url.ParseQuery(r.QueryString)
	if err != nil {
		return
	}

	var (
		urls           = query["url"]
		destenationURL = r.Header.Get("X-Callback-Url")
	)

	if urls == nil {
		log.Println("missing url parameter")
		return
	}
	for _, url := range urls {
		entries = append(entries, collectEntries(url)...)
	}

	raw, err := json.Marshal(entries)
	if err != nil {
		return
	}
	if destenationURL != "" {
		log.Printf("using callback %s\n", destenationURL)
		if err != nil {
			return
		}
		httpResponse, err = http.Post(destenationURL, "application/json", bytes.NewBuffer(raw))
		if err != nil {
			return
		}
		log.Printf("received x-callback-url %s response: %v\n", destenationURL, httpResponse)
	}

	response = handler.Response{
		Body:       raw,
		StatusCode: http.StatusOK,
		Header:     r.Header,
	}
	return
}

func collectEntries(url string) (entries []otodom.Entry) {

	c := colly.NewCollector()

	c.OnHTML("article[id]", func(e *colly.HTMLElement) {
		entry := otodom.Entry{
			Title:      e.ChildText("div.offer-item-details > header > h3 > a > span > span"),
			Name:       e.ChildText("div.offer-item-details-bottom > ul > li.pull-right"),
			Region:     e.ChildText("div.offer-item-details > header > p"),
			Price:      e.ChildText("div.offer-item-details > ul > li.hidden-xs.offer-item-price-per-m"),
			TotalPrice: e.ChildText("div.offer-item-details > ul > li.offer-item-price"),
			Area:       e.ChildText("div.offer-item-details > ul > li.hidden-xs.offer-item-area"),
			Link:       e.ChildAttr("div.offer-item-details > header > h3 > a", "href"),
		}
		entries = append(entries, entry)
	})

	c.OnRequest(func(r *colly.Request) {
		log.Println("visiting", r.URL.String())
	})

	c.Visit(url)

	log.Printf("collected %d records for url %s\n", len(entries), url)
	return entries
}