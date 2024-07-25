package main

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"log"
	"net/http"
	neturl "net/url"
	"strings"
	"sync"
	"time"
)

type Target struct {
	name    string
	apiUrl  string
	type_   PlatformType
	handler func(string, Target) ([]FilmInfo, string)
}

type FilmInfo struct {
	Name  string `json:"name"`
	Url   string `json:"url"`
	Img   string `json:"img"`
	Other string `json:"other"`
}

type FilmList struct {
	Name      string       `json:"name"`
	Films     []FilmInfo   `json:"films"`
	SearchUrl string       `json:"searchUrl"`
	Type      PlatformType `json:"type"`
}

type PlatformType int

const (
	Stream PlatformType = iota
	Download
)

func (st PlatformType) String() string {
	return [...]string{"Stream", "Download"}[st]
}
func (st PlatformType) MarshalJSON() ([]byte, error) {
	return json.Marshal(st.String())
}

var target = []Target{
	{"音范丝", "https://www.yinfans.me/?s=${word}", Download, yinfansiHandler},
	{"布谷TV", "https://www.bugutv.org/?cat=&s=${word}", Download, bugutvHandler},
	//{"茶杯狐", "https://cupfoxcc.com/chvodsearch/-------------.html?wd=${word}", Stream, cupfoxHandler},
	{"网飞猫", "https://www.ncat22.com/search?os=pc&k=${word}", Stream, ncatHandler},
	{"Vidhub", "https://vidhub.icu/search/${word}/", Stream, vidhubHandler},
}
var defaultHeaders = map[string]string{
	"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36",
}

func Search(word string) []FilmList {
	fmt.Println("Searching for:", word)
	var filmList []FilmList
	var wg sync.WaitGroup
	// 创建一个 channel 用于收集结果
	results := make(chan FilmList, len(target))
	for _, t := range target {
		wg.Add(1)
		go func(t Target) {
			defer wg.Done()
			filmInfos, url := t.handler(word, t)
			results <- FilmList{Films: filmInfos, Name: t.name, SearchUrl: url, Type: t.type_}
		}(t)
	}
	// 关闭 channel 并等待所有 goroutine 完成
	go func() {
		wg.Wait()
		close(results)
	}()

	// 收集所有结果
	for item := range results {
		filmList = append(filmList, item)
	}
	return filmList
}

func sendRequest(method string, url string, header map[string]string, reqBody io.Reader) ([]byte, *http.Response) {
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, nil
	}
	for s := range defaultHeaders {
		req.Header.Set(s, defaultHeaders[s])
	}
	if header != nil {
		for s := range header {
			req.Header.Set(s, header[s])
		}
	}

	// 创建一个HTTP客户端并发送请求
	client := http.Client{
		Timeout: 3 * time.Second,
		// 设置重定向策略
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return http.ErrUseLastResponse
			}
			req.Header.Del("Referer")
			return nil
		},
	}
	resp, err := client.Do(req)
	if resp == nil {
		log.Printf("resp is nil \n")
		return nil, nil
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)
	if err != nil || resp.StatusCode >= 400 {
		log.Printf("status: %d, Error sending request: %s \n", resp.StatusCode, err)
		return nil, resp
	}
	all, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %s \n", err)
		return nil, resp
	}

	// 输出响应
	//fmt.Println("Response Status:", resp.Status)
	return all, resp
}

func yinfansiHandler(word string, t Target) ([]FilmInfo, string) {
	reqUrl := strings.Replace(t.apiUrl, "${word}", word, 1)
	headers := map[string]string{
		"Referer": "https://www.yinfans.me/",
		"Host":    "www.yinfans.me",
		"Cookie":  "Hm_lvt_6c357a02991c9746bd18054c7da7d312=1719669488; esc_search_captcha=1; result=77; HMACCOUNT=CDDE847422F67B40; Hm_lpvt_6c357a02991c9746bd18054c7da7d312=1720145795; ppwp_wp_session=31343a25c8ff878e402c077f6cde4a32%7C%7C1721384759%7C%7C1721384399; wp_xh_session_02fbb834757a2600b9754229ee69d35c=31cb914279810c82f5f7c92b9cf78154%7C%7C1721555759%7C%7C1721552159%7C%7C38210b16cf2184ef24fdd2135e2da269",
	}
	body, _ := sendRequest("GET", reqUrl, headers, nil)
	if body == nil {
		return []FilmInfo{}, reqUrl
	}
	html, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		log.Fatal(err)
		return []FilmInfo{}, reqUrl
	}
	postList := html.Find("#post_container > .post")
	var filmList = make([]FilmInfo, 0)
	postList.Each(func(i int, s *goquery.Selection) {
		a := s.Find("a")
		title := a.AttrOr("title", "")
		url := a.AttrOr("href", "")
		img := a.Find("img").AttrOr("src", "")
		filmInfo := FilmInfo{Name: title, Url: url, Img: img}
		//fmt.Printf("%s", filmInfo)
		filmList = append(filmList, filmInfo)
	})
	return filmList, reqUrl
}

func bugutvHandler(word string, t Target) ([]FilmInfo, string) {
	reqUrl := strings.Replace(t.apiUrl, "${word}", word, 1)
	log.Print("reqUrl: ", reqUrl)
	headers := map[string]string{
		"Referer": "https://www.bugutv.org/",
		"Host":    "www.bugutv.org",
	}
	body, _ := sendRequest("GET", reqUrl, headers, nil)
	if body == nil {
		return []FilmInfo{}, reqUrl
	}
	html, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		log.Fatal(err)
		return []FilmInfo{}, reqUrl
	}
	postList := html.Find(".posts-wrapper article.post")
	var filmList = make([]FilmInfo, 0)
	postList.Each(func(i int, s *goquery.Selection) {
		a := s.Find(".entry-media a")
		title := a.AttrOr("title", "")
		url := a.AttrOr("href", "")
		img := a.Find("img").AttrOr("data-src", "")
		filmInfo := FilmInfo{Name: title, Url: url, Img: img}
		//fmt.Printf("%s", filmInfo)
		filmList = append(filmList, filmInfo)
	})
	//log.Printf("filmList: %v", filmList)
	return filmList, reqUrl
}

func ncatHandler(word string, t Target) ([]FilmInfo, string) {
	word = neturl.QueryEscape(word)
	reqUrl := strings.Replace(t.apiUrl, "${word}", word, 1)
	log.Print("reqUrl: ", reqUrl)
	headers := map[string]string{
		"Referer": "https://www.ncat22.com/",
		"Host":    "www.ncat22.com",
	}
	body, _ := sendRequest("GET", reqUrl, headers, nil)
	if body == nil {
		return []FilmInfo{}, reqUrl
	}
	html, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		log.Fatal(err)
		return []FilmInfo{}, reqUrl
	}
	postList := html.Find(".search-result-list a.search-result-item")
	var filmList = make([]FilmInfo, 0)
	postList.Each(func(i int, s *goquery.Selection) {
		a := s
		url := a.AttrOr("href", "")
		url = "https://www.ncat22.com" + url
		imgDom := s.Find(".search-result-item-pic > img")
		original := imgDom.AttrOr("data-original", "")
		img := imgDom.AttrOr("src", "")
		imgUrl, _ := neturl.Parse(img)
		newImg := imgUrl.Scheme + "://" + strings.Replace(imgUrl.Host, "15001", "15002", 1) + original
		title := imgDom.AttrOr("alt", "")
		filmInfo := FilmInfo{Name: title, Url: url, Img: newImg}
		//fmt.Printf("%s", filmInfo)
		filmList = append(filmList, filmInfo)
	})
	//log.Printf("filmList: %v", filmList)
	return filmList, reqUrl
}

func vidhubHandler(word string, t Target) ([]FilmInfo, string) {
	reqUrl := strings.Replace(t.apiUrl, "${word}", word, 1)
	log.Print("reqUrl: ", reqUrl)
	headers := map[string]string{
		"Referer": "https://vidhub.icu/",
		"Host":    "vidhub.icu",
	}
	body, _ := sendRequest("GET", reqUrl, headers, nil)
	if body == nil {
		return []FilmInfo{}, reqUrl
	}
	html, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		log.Fatal(err)
		return []FilmInfo{}, reqUrl
	}
	postList := html.Find(".video-img-box > .img-box")
	var filmList = make([]FilmInfo, 0)
	postList.Each(func(i int, s *goquery.Selection) {
		a := s.Find("a")
		url := a.AttrOr("href", "")
		imgDom := a.Find("img")
		img := imgDom.AttrOr("data-src", "")
		title := imgDom.AttrOr("alt", "")
		filmInfo := FilmInfo{Name: title, Url: url, Img: img}
		//fmt.Printf("%s", filmInfo)
		filmList = append(filmList, filmInfo)
	})
	//log.Printf("filmList: %v", filmList)
	return filmList, reqUrl
}

//func main() {
//	filmInfos := ncatHandler("流浪地球", target[2])
//	log.Printf("filmList: %v", filmInfos)
//}
