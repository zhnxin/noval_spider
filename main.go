package main

import (
	"noval_spider/core"
	"noval_spider/gui"
	"os"
	"os/user"
	"path"
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
	CONF           *AppConfig = &AppConfig{}
	CONF_FILE_PATH            = ".noval_spider_conf.toml"
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

func loadConfigFile(name string) (*AppConfig, error) {
	u, err := user.Current()
	if err != nil {
		return nil, err
	}
	file, err := os.OpenFile(path.Join(u.HomeDir, name), os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	conf := &AppConfig{}
	_, err = toml.DecodeReader(file, conf)
	return conf, err
}

func saveConfile(conf *AppConfig, name string) error {
	u, err := user.Current()
	if err != nil {
		return err
	}
	file, err := os.OpenFile(path.Join(u.HomeDir, name), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	err = toml.NewEncoder(file).Encode(conf)
	return err
}

func init() {
}

func main() {
	a := app.NewWithID("zhengxin.spider")
	a.SetIcon(theme.FyneLogo())
	w := a.NewWindow("spider")
	w.Resize(fyne.Size{Width: 400, Height: 300})
	w.SetMaster()
	taskM, container := gui.NewContainer(w)
	backToMainContentChan := make(chan struct{})
	configContainer := gui.NewConfigContainer(nil, nil)
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
		var err error
		CONF, err = loadConfigFile(CONF_FILE_PATH)
		if err != nil {
			errorExitProcess(a, w, err)
			return
		}
		if CONF.Level == "" {
			CONF.Level = "INFO"
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
		err := saveConfile(CONF, CONF_FILE_PATH)
		if err != nil {
			logrus.Error(err)
		}
	}()
	backToMainContentChan <- struct{}{}
	w.ShowAndRun()
}
