package main

import (
	"fmt"
	"noval_spider/core"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/theme"
)

func main() {
	a := app.NewWithID("zhengxin.spider")
	a.SetIcon(theme.FyneLogo())
	w := a.NewWindow("spider")
	w.SetMaster()
	w.SetMainMenu(fyne.NewMainMenu(fyne.NewMenu("File",
		fyne.NewMenuItem("New", func() { fmt.Println("Menu New") }),
	)))
	w.Resize(fyne.Size{Width: 400, Height: 300})
	taskM, container := NewContainer()
	conf := core.NewConfig("/YingShiShiJieDangShenTan/3254952.html", "YingShiShiJieDangShenTan.txt", false)
	conf.InjectDefault("https://www.soshuw.com", &core.ValidNext{EndWith: "html"}, &core.CssSelector{Title: "div.read_title>h1", Content: "div.content", Next: "div.pagego>a:last-child"})
	taskM.Add(conf)
	w.SetContent(container)
	w.ShowAndRun()
}
