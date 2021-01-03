package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

type (
	Config struct {
		OutFilePath     string
		TitleSelector   string
		ContentSelector string
		BaseUrl         string
		Proxy           string
		PageList        []string
	}
)

var (
	ReqClient = &http.Client{
		Timeout: time.Second * 5,
	}
)

func getConfig() *Config {
	outPutFilePath := kingpin.Flag("out", "out put file path").Required().Short('o').String()
	titleSelector := kingpin.Flag("title", "title").Short('t').String()
	contentSelector := kingpin.Flag("content", "content").Short('c').String()
	baseUrl := kingpin.Flag("base-url", "base url").Short('b').String()
	proxyUrl := kingpin.Flag("proxy", "proxy").Short('p').String()
	configPath := kingpin.Flag("config", "config").String()
	pageList := kingpin.Arg("url", "url").Strings()
	kingpin.Parse()
	c := &Config{}
	if *configPath != "" {
		_, err := toml.DecodeFile(*configPath, c)
		if err != nil {
			logrus.Fatal(err)
		}
	}
	if *titleSelector != "" {
		c.TitleSelector = *titleSelector
	}
	if *proxyUrl != "" {
		c.Proxy = *proxyUrl
	}
	if *contentSelector != "" {
		c.ContentSelector = *contentSelector
	}
	if *baseUrl != "" {
		c.BaseUrl = *baseUrl
	}
	if *outPutFilePath != "" {
		c.OutFilePath = *outPutFilePath
	}
	if len(*pageList) > 0 {
		c.PageList = *pageList
	}

	if c.ContentSelector == "" {
		logrus.Fatal("ContentSelector is required")
	}
	if c.Proxy != "" {
		proxyurl, err := url.Parse(c.Proxy)
		if err != nil {
			logrus.Fatalf("proxy url parse fail: %v", err)
		}
		ReqClient.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyurl),
		}
	}
	return c
}

func (c *Config) reqeust(url string, writer io.Writer) (err error) {
	resp, err := ReqClient.Get(url)
	if err != nil {
		return fmt.Errorf("request error:%v", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("status code:%d", resp.StatusCode)
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return err
	}
	if writer != nil {
		if c.TitleSelector != "" {
			doc.Find(c.TitleSelector).Each(func(i int, s *goquery.Selection) {
				_, err = writer.Write([]byte(s.Text()))
			})
			if err != nil {
				return err
			}

			if _, err = writer.Write([]byte{'\n'}); err != nil {
				return err
			}
		}
		doc.Find(c.ContentSelector).Each(func(i int, s *goquery.Selection) {
			_, err = writer.Write([]byte(s.Text()))
		})
		if err != nil {
			return err
		}
		if _, err = writer.Write([]byte{'\n'}); err != nil {
			return err
		}
	}
	return
}

func main() {
	c := getConfig()
	output, err := os.OpenFile(c.OutFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		logrus.Fatalf("open output file %s:%v", c.OutFilePath, err)
	}
	defer output.Close()
	for _, u := range c.PageList {
		if err = c.reqeust(c.BaseUrl+u, output); err != nil {
			logrus.Fatalf("open output file %s:%v", c.OutFilePath, err)
		}
	}

}
