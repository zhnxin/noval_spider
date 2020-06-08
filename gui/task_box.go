package gui

import (
	"fmt"
	"noval_spider/core"
	"sync"

	"fyne.io/fyne"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"github.com/sirupsen/logrus"
)

type (
	TaskEditCallBack func(func(*core.BaseConfig), *core.BaseConfig)
	TaskContainer    struct {
		editFunc  TaskEditCallBack
		container *widget.Box
		taskBox   []*TaskBox
		lock      *sync.Mutex
		window    fyne.Window
	}
	TaskBox struct {
		containerId int
		container   *TaskContainer

		isRunning   bool
		startOrStop *widget.Button
		nameLabel   *widget.Label
		statusLabel *widget.Label

		config *core.SpiderConfig
	}
)

func NewContainer(w fyne.Window) (*TaskContainer, fyne.CanvasObject) {
	container := &TaskContainer{
		container: widget.NewVBox(),
		taskBox:   []*TaskBox{},
		lock:      new(sync.Mutex),
		window:    w,
	}
	return container, widget.NewHScrollContainer(container.container)
}
func (c *TaskContainer) SetEditTaskFunc(fn TaskEditCallBack) {
	c.editFunc = fn
}
func (c *TaskContainer) GetAllConf() []core.BaseConfig {
	confs := make([]core.BaseConfig, len(c.taskBox))
	for i, t := range c.taskBox {
		confs[i] = t.config.BaseConfig
	}
	return confs
}
func (c *TaskContainer) Add(config *core.SpiderConfig) {
	task := &TaskBox{
		container:   c,
		isRunning:   false,
		nameLabel:   widget.NewLabel(config.Output),
		statusLabel: widget.NewLabel(""),
		config:      config,
	}
	config.SetLog(func(format string, a ...interface{}) {
		task.statusLabel.SetText(fmt.Sprintf(format, a...))
	})
	delBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		if task.isRunning {
			return
		}
		task.Remove()
	})
	delBtn.Enable()
	editBtn := widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {
		if task.isRunning || c.editFunc == nil {
			return
		} else {
			c.editFunc(func(conf *core.BaseConfig) {
				if conf == nil {
					return
				}
				logrus.Debugf("%+v", conf)
				task.config.InjectDefault(conf.Base, conf.ValidNext, conf.Selector)
				task.config.BaseConfig.Base = conf.Base
				task.config.BaseConfig.Start = conf.Start
				task.config.BaseConfig.Output = conf.Output
				task.config.BaseConfig.IsNext = conf.IsNext
				task.nameLabel.SetText(conf.Output)
			}, &task.config.BaseConfig)
		}

	})
	editBtn.Enable()
	task.startOrStop = widget.NewButtonWithIcon("", theme.MediaPlayIcon(), func() {
		if task.isRunning {
			task.config.Stop()
		} else {
			delBtn.Disable()
			editBtn.Disable()
			task.isRunning = true
			task.startOrStop.SetIcon(theme.MediaPauseIcon())
			task.startOrStop.Refresh()
			go func() {
				defer func() {
					delBtn.Enable()
					editBtn.Enable()
					task.isRunning = false
					task.startOrStop.SetIcon(theme.MediaPlayIcon())
					task.startOrStop.Refresh()
				}()
				err := task.config.Process()
				if err != nil {
					dialog.ShowError(err, c.window)
				}
			}()
		}
	})
	c.lock.Lock()
	defer c.lock.Unlock()
	task.containerId = len(c.taskBox)
	c.taskBox = append(c.taskBox, task)
	c.container.Append(widget.NewHBox(task.startOrStop, editBtn, delBtn, widget.NewVBox(task.nameLabel, task.statusLabel)))
	c.container.Refresh()
}

func (c *TaskContainer) Remove(id int) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if id >= len(c.taskBox) {
		return
	}
	for i := id; i < len(c.taskBox); i++ {
		c.taskBox[i].containerId--
	}
	c.taskBox = append(c.taskBox[0:id], c.taskBox[id+1:]...)
	c.container.Children = append(c.container.Children[0:id], c.container.Children[id+1:]...)
	c.container.Refresh()
}

func (b *TaskBox) Remove() {
	if b == nil {
		return
	}
	if b.container == nil {
		return
	}
	b.container.Remove(b.containerId)
}
