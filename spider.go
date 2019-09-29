package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

// Crawl posts from CSDN

const (
	ListPostURL   = "https://blog.csdn.net/%s/article/list/%d?"
	PostDetailURL = "https://mp.csdn.net/mdeditor/getArticle?id=%s"
)

type DetailData struct {
	Data PostDetail `json:"data"`
}

type PostDetail struct {
	Title string `json:"title"`
	Description string `json:"description"`
	Markdowncontent string `json:"markdowncontent"`
	Tags string `json:"tags"`
}

func GetPageSize(username string) (int, error) {
	client := http.Client{}

	resp, err := client.Get(fmt.Sprintf(ListPostURL, username, 1))
	if err != nil {
		return 0,err
	}

	data, err := ioutil.ReadAll(resp.Body)

	r := regexp.MustCompile(`class="ui-pager">.*?</li>`)
	finds := r.FindAll(data, -1)

	for _,f := range finds {
		ss := strings.Split(string(f), `<`)
		fmt.Println(ss)
	}

	return 0, nil
}

// Crawl posts by username
func CrawlPosts(username string, page int) ([]string, error) {
	client := http.Client{}

	resp, err := client.Get(fmt.Sprintf(ListPostURL, username, page))
	if err != nil {
		return nil,err
	}

	data, err := ioutil.ReadAll(resp.Body)

	r := regexp.MustCompile(`<h4 class="">\s*<a href=".*?"`)
	finds := r.FindAll(data, -1)

	var urls []string

	for _,f := range finds {
		ss := strings.Split(string(f), `"`)
		if len(ss) >= 4 {
			urls = append(urls, ss[3])
		}
	}

	return urls,err
}

func CrawlPostMarkdown(url string) (*PostDetail, error){

	index := strings.LastIndex(url, "/")
	id := url[index+1:]

	client := http.Client{}

	req, _ := http.NewRequest("GET", fmt.Sprintf(PostDetailURL, id), nil)
	req.Header.Set("cookie","uuid_tt_dd=10_33227520360-1562155374449-785950; UN=junmoxi; Hm_ct_6bcd52f51e9b3dce32bec4a3997715ac=6525*1*10_33227520360-1562155374449-785950!5744*1*junmoxi!1788*1*PC_VC; smidV2=20190705154448794d4aea42482882ccb01b435d4655850093278d5d0bb12e0; OUTFOX_SEARCH_USER_ID_NCOO=1275289703.8182168; dc_session_id=10_1565764323161.169173; UserName=junmoxi; UserInfo=de709e85392f4b8a8d19d69eb2273c56; UserToken=de709e85392f4b8a8d19d69eb2273c56; UserNick=java%E6%B4%BE%E5%A4%A7%E6%98%9F; AU=B09; BT=1567597499382; p_uid=U000000; notice=1; Hm_lvt_6bcd52f51e9b3dce32bec4a3997715ac=1569480050,1569545487,1569720826,1569734799; Hm_lpvt_6bcd52f51e9b3dce32bec4a3")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	detail := new(DetailData)
	err = json.Unmarshal(data, detail)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(data))

	fmt.Printf("%+v \n", detail)

	return nil, nil
}

func main() {
	//urls, err := CrawlPosts("junmoxi", 1)
	//if err != nil {
	//	panic(err)
	//}
	//
	//for _,url := range urls{
	//	fmt.Print(url)
	//}

	CrawlPostMarkdown("https://blog.csdn.net/junmoxi/article/details/101631412")

	// GetPageSize("junmoxi")
}