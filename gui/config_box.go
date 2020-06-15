package gui

import (
	"noval_spider/core"
	"os/user"
	"path"

	"fyne.io/fyne/widget"
	"github.com/sirupsen/logrus"
)

type (
	callBack              func(conf *core.BaseConfig)
	SpiderConfigContainer struct {
		submitFunc callBack
		cancelFunc func()

		baseInput   *widget.Entry
		startInput  *widget.Entry
		outputInput *widget.Entry
		isNext      *widget.Check

		validNextEndwith   *widget.Entry
		validNextNoContain *widget.Entry

		selectorTitle   *widget.Entry
		selectorContent *widget.Entry
		selectorNext    *widget.Entry

		confForm   *widget.Form
		spiderForm *widget.Form
	}
)

func NewConfigContainer(onsubmit callBack, oncancel func()) *SpiderConfigContainer {
	cantainer := &SpiderConfigContainer{submitFunc: onsubmit, cancelFunc: oncancel}
	cantainer.baseInput = widget.NewEntry()
	cantainer.startInput = widget.NewEntry()
	cantainer.outputInput = widget.NewEntry()
	cantainer.isNext = widget.NewCheck("isNext", func(bool) {})
	cantainer.validNextEndwith = widget.NewEntry()
	cantainer.validNextNoContain = widget.NewEntry()
	cantainer.selectorTitle = widget.NewEntry()
	cantainer.selectorContent = widget.NewEntry()
	cantainer.selectorNext = widget.NewEntry()
	return cantainer
}

func (c *SpiderConfigContainer) SetOnSubmit(fn callBack) {
	c.submitFunc = fn
	if c.confForm != nil {
		c.confForm.OnSubmit = func() {
			fn(c.getConf())
		}
	}
	if c.spiderForm != nil {
		c.spiderForm.OnSubmit = func() {
			fn(c.getConf())
		}
	}
}
func (c *SpiderConfigContainer) SetOnCannel(fn func()) {
	c.cancelFunc = fn
	if c.confForm != nil {
		c.confForm.OnCancel = func() { c.cancelFunc(); c.submitFunc(nil) }
	}
	if c.spiderForm != nil {
		c.spiderForm.OnCancel = func() { c.cancelFunc(); c.submitFunc(nil) }
	}
}
func (c *SpiderConfigContainer) getConf() *core.BaseConfig {
	conf := core.NewBaseConfig(c.startInput.Text, c.outputInput.Text, c.isNext.Checked)
	conf.Base = c.baseInput.Text
	if c.validNextEndwith.Text != "" || c.validNextNoContain.Text != "" {
		conf.ValidNext = &core.ValidNext{EndWith: c.validNextEndwith.Text, NotContains: c.validNextNoContain.Text}
	}
	if c.selectorContent.Text != "" || c.selectorTitle.Text != "" || c.selectorNext.Text != "" {
		conf.Selector = &core.CssSelector{
			Content: c.selectorContent.Text,
			Title:   c.selectorTitle.Text,
			Next:    c.selectorNext.Text,
		}
	}
	return &conf
}

func (c *SpiderConfigContainer) restore(conf *core.BaseConfig) {
	c.baseInput.SetText(conf.Base)
	if conf.Output == "" {
		u, err := user.Current()
		if err == nil {
			logrus.Info(path.Join(u.HomeDir, "Downloads"))
			c.outputInput.SetText(path.Join(u.HomeDir, "Downloads"))
		}
	} else {
		c.outputInput.SetText(conf.Output)
	}
	c.isNext.SetChecked(conf.IsNext)
	c.startInput.SetText(conf.Start)

	if conf.ValidNext != nil {
		c.validNextEndwith.SetText(conf.ValidNext.EndWith)
		c.validNextNoContain.SetText(conf.ValidNext.NotContains)
	}
	if conf.Selector != nil {
		c.selectorTitle.SetText(conf.Selector.Title)
		c.selectorContent.SetText(conf.Selector.Content)
		c.selectorNext.SetText(conf.Selector.Next)
	}
}

func (c *SpiderConfigContainer) NewConfigForm(conf *core.BaseConfig) *widget.Form {
	c.restore(conf)
	// if c.confForm != nil {
	// 	return c.confForm
	// }
	c.confForm = &widget.Form{
		OnCancel: c.cancelFunc,
		OnSubmit: func() {
			c.submitFunc(c.getConf())
		},
	}
	c.confForm.Append("Base", c.baseInput)
	c.confForm.Append("Title Css Selector", c.selectorTitle)
	c.confForm.Append("Content Css Selector", c.selectorContent)
	c.confForm.Append("Next Css Selector", c.selectorNext)
	c.confForm.Append("Valid Next EndWith", c.validNextEndwith)

	c.confForm.Append("Valid Next NoContain", c.validNextNoContain)
	c.confForm.Append("http proxy", c.startInput)
	return c.confForm
}
func (c *SpiderConfigContainer) NewSpiderConfigForm(conf *core.BaseConfig) *widget.Form {
	c.restore(conf)
	// if c.spiderForm != nil {
	// 	return c.confForm
	// }
	c.spiderForm = &widget.Form{
		OnCancel: c.cancelFunc,
		OnSubmit: func() {
			c.submitFunc(c.getConf())
		},
	}
	c.spiderForm.Append("Base", c.baseInput)
	c.spiderForm.Append("Output", c.outputInput)
	c.spiderForm.Append("Start", c.startInput)
	c.spiderForm.Append("Is Next", c.isNext)
	c.spiderForm.Append("Title Css Selector", c.selectorTitle)
	c.spiderForm.Append("Content Css Selector", c.selectorContent)
	c.spiderForm.Append("Next Css Selector", c.selectorNext)
	c.spiderForm.Append("Valid Next EndWith", c.validNextEndwith)

	c.spiderForm.Append("Valid Next NoContain", c.validNextNoContain)
	return c.spiderForm
}
