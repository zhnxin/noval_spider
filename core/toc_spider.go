package core

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
)

type (
	TocSpiderConfig struct {
		BaseConfig
		cannel context.CancelFunc
		log    LogFunc
	}
)

func (c *TocSpiderConfig) reqeust(url string, writer io.Writer) (err error) {
	res, _, errs := ReqAgent.Proxy(httpProxy).Get(url).End()
	if len(errs) > 0 {
		return fmt.Errorf("request error:%v", errs)
	}
	if res.StatusCode != 200 {
		return fmt.Errorf("status code:%d", res.StatusCode)
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return err
	}
	if writer != nil {
		c.CurrentChapter++
		doc.Find(c.Selector.Content).Each(func(i int, s *goquery.Selection) {
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

func (c *TocSpiderConfig) Stop() {
	if c.cannel != nil {
		c.cannel()
	}
}

func (c *TocSpiderConfig) Process() (err error) {
	if c.log == nil {
		c.log = func(format string, a ...interface{}) { logrus.Infof(format, a...) }
	}
	res, _, errs := ReqAgent.Proxy(httpProxy).Get(c.Base + c.Start).End()
	if len(errs) > 0 {
		return fmt.Errorf("request error:%v", errs)
	}
	if res.StatusCode != 200 {
		return fmt.Errorf("status code:%d", res.StatusCode)
	}
	c.log(c.Base + c.Start)
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return err
	}
	output, err := os.OpenFile(c.Output, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		return fmt.Errorf("open output file %s:%v", c.Output, err)
	}
	c.log(c.Selector.Title)
	doc.Find(c.Selector.Title).Each(func(i int, s *goquery.Selection) {
		c.log("title: %s", s.Text())
		if c.TitleDecorator != "" {
			_, err = fmt.Fprintf(output, c.TitleDecorator, i+1, s.Text())
		} else {
			_, err = output.Write([]byte(s.Text()))
		}
		if contentUrl, ok := s.Attr("href"); ok {
			if err := c.reqeust(c.Base+contentUrl, output); err != nil {
				c.log("get content:%v", err)
			}
		}
	})
	return nil
}
