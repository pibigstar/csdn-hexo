package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
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
	PostDetailURL = "https://bizapi.csdn.net/blog-console-api/v3/editor/getArticle?id=%s&model_type="
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
 	postTime = time.Now()
)

const (
	appSecret = "9znpamsyl2c7cdrr9sas0le9vbc3r6ba"
	appCaKey = "203803574"
	signHeaders = "x-ca-key,x-ca-nonce"
)

func init() {
	flag.StringVar(&username, "username", "junmoxi", "your csdn username")
	flag.StringVar(&cookie, "cookie", "UserName=junmoxi;UserToken=34543a5e65f7cae7cb3c4;", "your csdn cookie")
	flag.IntVar(&page, "page", -1, "download pages")
	flag.Parse()
	rand.Seed(time.Now().Unix())
}

func main() {
	urls, err := crawlPosts(username)
	if err != nil {
		panic(err)
	}
	bar = pgbar.NewBar(0, "下载进度", len(urls))

	for _, ul := range urls {
		wg.Add(1)
		go crawlPostMarkdown(ul)
	}
	wg.Wait()
}

// Crawl posts by username
func crawlPosts(username string) ([]string, error) {
	defer fmt.Println("地址抓取完成,开始下载...")

	var urls []string
	for {
		fmt.Printf("正在抓取第%d页文章地址... \n", currentPage)
		resp, err := http.DefaultClient.Get(fmt.Sprintf(ListPostURL, username, currentPage))
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
	defer wg.Done()
	defer bar.Add()
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	index := strings.LastIndex(url, "/")
	id := url[index+1:]
	apiUrl := fmt.Sprintf(PostDetailURL, id)

	uuid := createUUID()
	sign := createSignature(uuid, apiUrl)

	req, _ := http.NewRequest("GET",apiUrl, nil)
	req.Header.Set("cookie", cookie)
	req.Header.Set("x-ca-key", appCaKey)
	req.Header.Set("x-ca-nonce", uuid)
	req.Header.Set("x-ca-signature", sign)
	req.Header.Set("x-ca-signature-headers", signHeaders)
	req.Header.Set("Accept", "*/*")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		return
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	var post DetailData
	err = json.Unmarshal(data, &post)
	if err != nil {
		return
	}

	if post.Data.Markdowncontent != "" {
		 buildMarkdownPost(post.Data)
	} else if post.Data.Content != "" {
		 buildHtmlPost(post.Data)
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
}

func createSignature(uuid, apiUrl string) string {
	u, err := url.Parse(apiUrl)
	if err != nil {
		panic(err)
	}
	query := u.Query().Encode()
	query = query[:len(query)-1]
	message := fmt.Sprintf("GET\n*/*\n\n\n\nx-ca-key:%s\nx-ca-nonce:%s\n%s?%s", appCaKey, uuid, u.Path, query)
	hc := hmac.New(sha256.New, []byte(appSecret))
	hc.Write([]byte(message))
	res := hc.Sum(nil)

	result := base64.StdEncoding.EncodeToString(res)
	return result
}

func createUUID() string {
	s := strings.Builder{}
	chars := make([]string, 0, 10)
	for i := 97; i < 103; i++ {
		chars = append(chars, string(i))
	}
	for i := 49; i < 58; i++ {
		chars = append(chars, string(i))
	}
	xs := "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx"
	for _, k := range xs {
		x := string(k)
		if x == "4" || x == "-" {
			s.WriteString(x)
		} else {
			i := rand.Intn(len(chars))
			s.WriteString(chars[i])
		}
	}
	return s.String()
}