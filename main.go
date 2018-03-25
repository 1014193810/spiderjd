// spider project main.go
package main

import (
	"fmt"
	//	"os"
	"flag"

	"strconv"

	"io/ioutil"
	"net/http"
	"regexp"

	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/axgle/mahonia"
)

var (
	comment = regexp.MustCompile(`\"commentCount\":[0-9]+`)
	good    = regexp.MustCompile(`\"goodRate\":[0-9]\.[0-9]+`)
	mode    = flag.Int64("m", 1, "Usage:1-search in page ,2-search by every produc")
	keyword = flag.String("k", "iphone", "Usage: keyword")
	pnum    = flag.Int("n", 1, "the num of page")
)

type product struct {
	name    string
	price   string
	comment string
	shop    string
	icons   string
	src     string
}
type Products []product

func main() {
	flag.Parse()
	fmt.Println("mode:", *mode, "keyword:", *keyword, "pagenum:", *pnum)

	urls := SearchPages(*keyword, *pnum)
	if *mode == 2 {

		itemurls := Pull(urls)
		for itemurl, _ := range itemurls {
			id := itemurl[12 : len(itemurl)-5]
			Crawl(id)
		} //这一段逐个爬取单个产品信息，速度慢

	} else if *mode == 1 {
		Pull2(urls) //直接从搜索页面中爬取，速度快
	} else {
		fmt.Println("illegal mode")
	}

}
func Pull2(urls []string) *Products {
	var out Products
	for _, url := range urls {

		doc, err := goquery.NewDocument(url)
		if err != nil {
			fmt.Println(err)
		}
		glist := doc.Find("div.goods-list-v2")
		glist.Find("li.gl-item").Each(func(i int, s *goquery.Selection) {
			price := s.Find("div.p-price").Find("i").Text()
			name := s.Find("div.p-name").Find("em").Text()
			commit := s.Find("div.p-commit").Find("a").Text()
			shop := s.Find("div.p-shop").Find("a").Text()
			src, _ := s.Find("div.p-img").Find("a").Attr("href")
			icons := s.Find("div.p-icons").Find("i").Text()
			out = append(out, product{
				name:    name,
				price:   price,
				comment: commit,
				shop:    shop,
				icons:   icons,
				src:     src,
			})
			fmt.Println(i, price, name, commit, src, shop, icons)
		})

	}
	return &out
}

func Crawl(id string) {
	doc, err := goquery.NewDocument("https://item.m.jd.com/product/" + id + ".html")
	if err != nil {
		fmt.Printf(err.Error())
	}
	title := doc.Find("span.title-text").Text()
	price := doc.Find("div.prod-price").Text()
	p := Getcomment(id)
	fmt.Println(title, "价格："+strings.TrimSpace(price), p)
	return
}
func Getbody(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	out := changetype(string(body))
	return out
}
func changetype(src string) string {
	srcCoder := mahonia.NewDecoder("gbk")
	result := srcCoder.ConvertString(src)
	return result
}
func Getcomment(id string) string {
	body := Getbody("http://club.jd.com/productpage/p-" + id + "-s-0-t-3-p-0.html")
	out := "评论：" + comment.FindAllString(body, 1)[0][15:] + "  好评：" + good.FindAllString(body, 1)[0][11:]
	return out
}
func SearchPages(keyword string, pages int) []string {
	urls := []string{}
	for i := 0; i < pages; i++ {
		urli := "https://search.jd.com/Search?keyword=" + keyword + "&page=" + strconv.Itoa(i+i-1) + "&enc=utf-8"
		urls = append(urls, urli)
	}
	return urls
}
func Pull(urls []string) map[string]bool {

	itemreg := regexp.MustCompile(`item\.jd\.com/[0-9]+\.html`)
	out := make(map[string]bool)
	for _, url := range urls {

		body := Getbody(url)

		itemurls := itemreg.FindAllString(body, -1)
		for _, url := range itemurls {
			out[url] = true
		}

	}
	return out
}
