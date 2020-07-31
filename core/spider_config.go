package core

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/parnurzeal/gorequest"
	"github.com/sirupsen/logrus"
)

var (
	ReqAgent  = gorequest.New()
	httpProxy = ""
)

func SetProxy(httpProxyUrl string) {
	httpProxy = httpProxyUrl
}

func init() {
	ReqAgent.SetDoNotClearSuperAgent(true)
}

type (
	LogFunc    func(format string, a ...interface{})
	BaseConfig struct {
		Base           string
		Start          string
		IsNext         bool
		Output         string
		TitleDecorator string
		CurrentChapter int
		ValidNext      *ValidNext
		Selector       *CssSelector
	}
	SpiderConfig struct {
		BaseConfig
		cannel context.CancelFunc
		log    LogFunc
	}
	CssSelector struct {
		Title   string
		Content string
		Next    string
	}
	ValidNext struct {
		EndWith     string
		NotContains string
	}
)

func NewBaseConfig(start, output string, isNext bool) BaseConfig {
	return BaseConfig{
		Start:  start,
		Output: output,
		IsNext: isNext,
	}
}
func (c *BaseConfig) SpiderConfig() *SpiderConfig {
	return &SpiderConfig{
		BaseConfig: *c,
	}
}
func NewConfig(start, output string, isNext bool) *SpiderConfig {
	return &SpiderConfig{
		BaseConfig: NewBaseConfig(start, output, isNext),
	}
}

func (c *SpiderConfig) InjectDefault(base string, next *ValidNext, selector *CssSelector) {
	if c.Base == "" {
		c.Base = base
	}
	if c.ValidNext == nil {
		c.ValidNext = next
	}
	if c.Selector == nil {
		c.Selector = selector
	}
	if c.log == nil {
		c.log = func(format string, a ...interface{}) { logrus.Infof(format, a...) }
	}

}
func (c *SpiderConfig) SetLog(log LogFunc) {
	c.log = log
}

func (c *SpiderConfig) reqeust(url string, writer io.Writer) (next string, err error) {
	res, _, errs := ReqAgent.Proxy(httpProxy).Get(url).End()
	if len(errs) > 0 {
		return "", fmt.Errorf("request error:%v", errs)
	}
	if res.StatusCode != 200 {
		return "", fmt.Errorf("status code:%d", res.StatusCode)
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", err
	}
	if writer != nil {
		c.CurrentChapter++
		if c.Selector.Title != "" {
			doc.Find(c.Selector.Title).Each(func(i int, s *goquery.Selection) {
				if c.TitleDecorator != "" {
					_, err = fmt.Fprintf(writer, c.TitleDecorator, c.CurrentChapter, s.Text())
				} else {
					_, err = writer.Write([]byte(s.Text()))
				}
			})
			if err != nil {
				return "", err
			}

			if _, err = writer.Write([]byte{'\n'}); err != nil {
				return "", err
			}
		}
		doc.Find(c.Selector.Content).Each(func(i int, s *goquery.Selection) {
			_, err = writer.Write([]byte(s.Text()))
		})
		if err != nil {
			return "", err
		}
		if _, err = writer.Write([]byte{'\n'}); err != nil {
			return "", err
		}
	}
	next, _ = doc.Find(c.Selector.Next).Attr("href")
	logrus.Debug(c.Selector.Next, ": ", next)
	if c.ValidNext.NotContains != "" &&
		strings.Contains(next, c.ValidNext.NotContains) {
		next = ""
	}
	if c.ValidNext.EndWith != "" &&
		!strings.HasSuffix(next, c.ValidNext.EndWith) {
		next = ""
	}
	return
}
func (c *SpiderConfig) Stop() {
	if c.cannel != nil {
		c.cannel()
	}
}

func (c *SpiderConfig) Process() (err error) {
	var url string
	if c.IsNext {
		url, err = c.reqeust(c.Base+c.Start, nil)
		if err != nil {
			return
		}
		if url == "" {
			c.log("no next page for %s\n", c.Start)
			return
		}
		c.Start = url
		c.IsNext = false
	}
	output, err := os.OpenFile(c.Output, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		return fmt.Errorf("open output file %s:%v", c.Output, err)
	}
	defer output.Close()
	ctx, cannel := context.WithCancel(context.Background())
	c.cannel = cannel
	defer func() {
		c.cannel = nil
	}()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			url, err = c.reqeust(c.Base+c.Start, output)
			if err != nil {
				return err
			}
			if url == "" {
				c.IsNext = true
				return nil
			}
			c.log("next: %d   %s\n", c.CurrentChapter, url)
			c.Start = url
			c.IsNext = false
		}
	}
}
