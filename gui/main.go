package main

import (
	"noval_spider/core"
	"os"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"github.com/BurntSushi/toml"
	"github.com/sirupsen/logrus"
)

var (
	CONF *AppConfig = &AppConfig{}
)

type (
	AppConfig struct {
		Level   string
		Default core.BaseConfig
		Spider  []core.BaseConfig
	}
)

func errorExitProcess(a fyne.App, w fyne.Window, err error) {
	prog := dialog.NewProgress("Unexpected error", err.Error(), w)
	go func() {
		num := 0.0
		for num < 1.0 {
			prog.SetValue(1 - num)
			time.Sleep(time.Millisecond * 100)
			num += 0.05
		}
		prog.SetValue(0)
		time.Sleep(time.Millisecond * 50)
		prog.Hide()
		a.Quit()
	}()
	prog.Show()
}

func init() {
}

func main() {
	a := app.NewWithID("zhengxin.spider")
	a.SetIcon(theme.FyneLogo())
	w := a.NewWindow("spider")
	w.Resize(fyne.Size{Width: 400, Height: 300})
	w.SetMaster()
	taskM, container := NewContainer(w)
	// conf := core.NewConfig("/YingShiShiJieDangShenTan/3254952.html", "YingShiShiJieDangShenTan.txt", false)
	// conf.InjectDefault("https://www.soshuw.com", &core.ValidNext{EndWith: "html"}, &core.CssSelector{Title: "div.read_title>h1", Content: "div.content", Next: "div.pagego>a:last-child"})
	// taskM.Add(conf)
	backToMainContentChan := make(chan struct{})
	configContainer := NewConfigContainer(nil, nil)
	confBar := fyne.NewContainerWithLayout(layout.NewFixedGridLayout(fyne.NewSize(60, 48)),
		widget.NewButton("Quit", func() { logrus.Debugf("退出 button tapped"); a.Quit() }),
		widget.NewButtonWithIcon("", theme.FolderNewIcon(), func() {
			logrus.Debugf("新增 button Tapped")
			configContainer.SetOnSubmit(func(conf *core.BaseConfig) {
				spiderConf := conf.SpiderConfig()
				spiderConf.InjectDefault(CONF.Default.Base, CONF.Default.ValidNext, CONF.Default.Selector)
				taskM.Add(spiderConf)
				backToMainContentChan <- struct{}{}
			})
			w.SetContent(configContainer.NewSpiderConfigForm(&CONF.Default))
		}),
		widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {
			logrus.Debugf("配置 button tapped")
			configContainer.SetOnSubmit(func(conf *core.BaseConfig) {
				CONF.Default = *conf
				logrus.Debugf("更新配置:%+v", CONF.Default)
				backToMainContentChan <- struct{}{}
			})
			w.SetContent(configContainer.NewConfigForm(&CONF.Default))
		}),
	)
	go func() {
		for range backToMainContentChan {
			w.SetContent(fyne.NewContainerWithLayout(
				layout.NewBorderLayout(confBar, nil, nil, nil), confBar, container))
		}
	}()
	configContainer.SetOnCannel(func() {
		backToMainContentChan <- struct{}{}
	})
	go func() {
		file, err := os.OpenFile(".noval_spider_conf.toml", os.O_CREATE|os.O_RDONLY, 0644)
		if err != nil {
			dialog.ShowError(err, w)
			time.Sleep(5 * time.Second)
			a.Quit()
			return
		}
		defer file.Close()
		CONF.Level = "INFO"
		_, err = toml.DecodeReader(file, CONF)
		if err != nil {
			errorExitProcess(a, w, err)
			return
		}
		level, err := logrus.ParseLevel(CONF.Level)
		if err != nil {
			errorExitProcess(a, w, err)
			return
		}
		logrus.SetLevel(level)

		for _, c := range CONF.Spider {
			func(cf core.BaseConfig) {
				spiConf := cf.SpiderConfig()
				spiConf.InjectDefault(CONF.Default.Base, CONF.Default.ValidNext, CONF.Default.Selector)
				taskM.Add(spiConf)
			}(c)

		}
	}()
	defer func() {
		CONF.Spider = taskM.GetAllConf()
		file, err := os.OpenFile(".noval_spider_conf.toml", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			logrus.Error(err)
			return
		}
		defer file.Close()
		logrus.Debugf("%+v", CONF.Default)
		err = toml.NewEncoder(file).Encode(CONF)
		if err != nil {
			logrus.Error(err)
		}
	}()
	backToMainContentChan <- struct{}{}
	w.ShowAndRun()
}
