package main

import (
	"fmt"
	"noval_spider/core"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/BurntSushi/toml"

	"gopkg.in/alecthomas/kingpin.v2"
)

type (
	ConfigPackage struct {
		Base      string
		ValidNext core.ValidNext
		Selector  core.CssSelector
		Config    []core.SpiderConfig
	}
)

var (
	ConfigPath = kingpin.Flag("config", "config file for multi").Default("config.toml").Short('c').String()
	IsDebug    = kingpin.Flag("debug", "is print debug log").Bool()
	IsInit     = kingpin.Flag("init", "init the config file").Bool()
)

func (cp *ConfigPackage) ReOrg() {
	for i := 0; i < len(cp.Config); i++ {
		cp.Config[i].InjectDefault(cp.Base, &cp.ValidNext, &cp.Selector)
	}
}

func createConfigFile() error {
	f, err := os.OpenFile("config.toml", os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(f, `
base=""
[CssSelector]
Title=''
Content='div.box_box'
Next='#keyright'
[ValidNext]
EndWith=''
NotContains=''
[[Config]]
start=''
isnext = false
output="output.txt"`)
	return err
}

func main() {
	kingpin.Parse()
	if *IsInit {
		if err := createConfigFile(); err != nil {
			logrus.Fatalln(err)
		}
		fmt.Println("config.toml creation completed")
		return
	}
	CONFIG := ConfigPackage{}
	_, err := toml.DecodeFile(*ConfigPath, &CONFIG)
	if err != nil {
		logrus.Fatalln(err)
	}
	CONFIG.ReOrg()
	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: time.RFC3339,
		FullTimestamp:   true,
	})
	if *IsDebug {
		logrus.SetLevel(logrus.DebugLevel)
	}
	logrus.Debugf("config: %+v", CONFIG)

	wait := new(sync.WaitGroup)
	configChan := make(chan core.SpiderConfig)
	for _, c := range CONFIG.Config {
		wait.Add(1)
		go func(con core.SpiderConfig) {
			defer func() {
				wait.Done()
				configChan <- con
			}()
			if err := con.Process(); err != nil {
				logrus.Errorln(err)
			} else {
				logrus.Info("complete: ", con.Output)
			}
		}(c)
	}
	go func() {
		wait.Wait()
		close(configChan)
	}()
	CONFIG.Config = []core.SpiderConfig{}
	for c := range configChan {
		CONFIG.Config = append(CONFIG.Config, c)
	}
	confFile, err := os.OpenFile(*ConfigPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		logrus.Fatal("update config file: ", err)
	}
	err = toml.NewEncoder(confFile).Encode(CONFIG)
	if err != nil {
		logrus.Fatal("update config file: ", err)
	}
}
