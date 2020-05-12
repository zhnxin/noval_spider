package main

import (
	"fmt"
	"noval_spider/core"
	"sync"

	"fyne.io/fyne"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

type (
	TaskContainer struct {
		container *widget.Box
		taskBox   []*TaskBox
		lock      *sync.Mutex
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

func NewContainer() (*TaskContainer, fyne.CanvasObject) {
	container := &TaskContainer{
		container: widget.NewVBox(),
		taskBox:   []*TaskBox{},
		lock:      new(sync.Mutex),
	}
	return container, widget.NewHScrollContainer(container.container)
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
	task.startOrStop = widget.NewButtonWithIcon("", theme.MediaPlayIcon(), func() {
		if task.isRunning {
			task.config.Stop()
		} else {
			delBtn.Disable()
			task.isRunning = true
			task.startOrStop.SetIcon(theme.MediaPauseIcon())
			task.startOrStop.Refresh()
			//TODO deal with the error
			go func() {
				defer func() {
					delBtn.Enable()
					task.isRunning = false
					task.startOrStop.SetIcon(theme.MediaPlayIcon())
					task.startOrStop.Refresh()
				}()
				_ = task.config.Process()
			}()
		}
	})
	c.lock.Lock()
	defer c.lock.Unlock()
	task.containerId = len(c.taskBox)
	c.taskBox = append(c.taskBox, task)
	c.container.Append(widget.NewHBox(task.startOrStop, delBtn, widget.NewVBox(task.nameLabel, task.statusLabel)))
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
