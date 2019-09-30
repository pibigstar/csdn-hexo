package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/qianlnk/pgbar"
)

// Crawl posts from csdn
// build posts to hexo style

const (
	ListPostURL   = "https://blog.csdn.net/%s/article/list/%d?"
	PostDetailURL = "https://mp.csdn.net/mdeditor/getArticle?id=%s"
	HexoHeader    = `
---
title: %s
date: %s
tags: [%s]
categories: %s
---
`
	HtmlBody = `<html>
<head>
<title>%s</title>
</head>
<body>
%s
</body>
</html>`
)

var postTime = time.Now()

type DetailData struct {
	Data PostDetail `json:"data"`
}

type PostDetail struct {
	Title           string `json:"title"`
	Description     string `json:"description"`
	Content         string `json:"content"`
	Markdowncontent string `json:"markdowncontent"`
	Tags            string `json:"tags"`
	Categories      string `json:"categories"`
}

var (
	username    string
	page        int
	cookie      string
	currentPage = 1
	count       int
	wg          sync.WaitGroup
	bar         *pgbar.Bar
)

func init() {
	flag.StringVar(&username, "username", "junmoxi", "your csdn username")
	flag.StringVar(&cookie, "cookie", "UserName=junmoxi; UserToken=c3c29cca48be43c4884fe36d052d5851;", "your csdn cookie")
	flag.IntVar(&page, "page", -1, "download pages")
	flag.Parse()
}

func main() {
	urls, err := crawlPosts(username)
	if err != nil {
		panic(err)
	}

	bar = pgbar.NewBar(0, "下载进度", len(urls))
	for _, url := range urls {
		wg.Add(1)
		go crawlPostMarkdown(url)
	}

	wg.Wait()
}

// Crawl posts by username
func crawlPosts(username string) ([]string, error) {
	client := http.Client{}
	var (
		urls []string
	)

	for {
		resp, err := client.Get(fmt.Sprintf(ListPostURL, username, currentPage))
		if err != nil {
			return nil, err
		}

		data, err := ioutil.ReadAll(resp.Body)

		r := regexp.MustCompile(`<h4 class="">\s*<a href=".*?"`)
		finds := r.FindAll(data, -1)

		for _, f := range finds {
			ss := strings.Split(string(f), `"`)
			if len(ss) >= 4 {
				urls = append(urls, ss[3])
			}
		}

		if len(finds) == 0 {
			return urls, nil
		}

		if page != -1 && currentPage >= page {
			return urls, nil
		}
		currentPage++
	}
}

func crawlPostMarkdown(url string) {
	index := strings.LastIndex(url, "/")
	id := url[index+1:]

	client := http.Client{}

	req, _ := http.NewRequest("GET", fmt.Sprintf(PostDetailURL, id), nil)
	req.Header.Set("cookie", cookie)

	resp, err := client.Do(req)
	if err != nil {
		return
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	post := new(DetailData)
	err = json.Unmarshal(data, post)
	if err != nil {
		return
	}

	if post.Data.Markdowncontent != "" {
		go buildMarkdownPost(post.Data)
	} else if post.Data.Content != "" {
		go buildHtmlPost(post.Data)
	}
}

func buildMarkdownPost(post PostDetail) {
	date := postTime.Format("2006-01-02 15:03:04")
	header := fmt.Sprintf(HexoHeader, post.Title, date, post.Tags, post.Categories)

	err := ioutil.WriteFile(
		fmt.Sprintf("%s.md", post.Title),
		[]byte(fmt.Sprintf("%s\n%s", header, post.Markdowncontent)),
		os.ModePerm)

	if err != nil {
		return
	}

	rand.Seed(time.Now().UnixNano())
	d := rand.Intn(3) + 1
	postTime = postTime.AddDate(0, 0, -d).Add(time.Hour)

	count++

	defer wg.Done()
	defer bar.Add()
}

func buildHtmlPost(post PostDetail) {

	html := fmt.Sprintf(HtmlBody, post.Title, post.Content)
	err := ioutil.WriteFile(
		fmt.Sprintf("%s.html", post.Title),
		[]byte(fmt.Sprintf("%s", html)),
		os.ModePerm)
	if err != nil {
		return
	}

	defer wg.Done()
	defer bar.Add()
}
