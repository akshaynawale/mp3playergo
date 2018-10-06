package main

import (
	"flag"
	"fmt"
	"github.com/akshaynawale/gorpisense/joyst"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	ui "github.com/gizak/termui"
	"github.com/golang/glog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var soundpath string

func init() {
	flag.StringVar(&soundpath, "d", "./", "Provide path to the mp3 sound files")
	flag.Set("logtostderr", "true")
}

func (a *audios) playSound() {
	for {
		fileName := fmt.Sprintf("./%s", a.audioList[a.playing])
		a.playfile(fileName)
		<-a.cursor
	}
}

func (a *audios) playfile(fileName string) {
	file, err := os.Open(fileName)
	if err != nil {
		glog.Fatal(err)
	}
	glog.Infof("Now Playing %s", fileName)
	a.updateUi()
	//a.NowPlaying <- fileName
	s, format, _ := mp3.Decode(file)
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	speaker.Play(beep.Seq(s, beep.Callback(func() {
		if a.playing == 0 {
			a.playing = len(a.audioList) - 1
		} else {
			a.playing = a.playing - 1
		}
		a.cursor <- a.playing
	})))
}

type audios struct {
	audioList []string // List of songss availble
	playing   int      // now playing songs
	cursor    chan int // chan to manage changing songs
	Stop      chan int
	//NowPlaying chan string // Name of the song playing
	par0  *ui.Par
	gauge *ui.Gauge
	vol   int
}

func (a *audios) Init() {
	files, err := filepath.Glob(soundpath)
	if err != nil {
		glog.Fatal(err)
	}
	glog.Infof("Songs found: %s ", files)
	a.audioList = files
	a.playing = 0
	a.cursor = make(chan int)
	///a.cursor <- 0
	glog.Infof("Now playing set to %d", a.playing)
	a.Stop = make(chan int)
	//a.NowPlaying = make(chan string)
	a.vol = 50
}

func (a *audios) displayInit() {
	err := ui.Init()
	if err != nil {
		glog.Fatal(err)
	}
	ls := ui.NewList()
	ls.Items = a.audioList
	ls.ItemFgColor = ui.ColorYellow
	ls.BorderLabel = "Audio Files:"
	ls.Height = 7
	ls.Width = 25
	ls.Y = 0

	a.par0 = ui.NewPar(a.audioList[a.playing])
	a.par0.Height = 3
	a.par0.Width = 20
	a.par0.Y = 1
	a.par0.BorderLabel = "Now Playing"

	a.gauge = ui.NewGauge()
	a.gauge.Percent = 40
	a.gauge.Width = 2
	a.gauge.Height = 3
	a.gauge.BorderLabel = "Volume"
	a.gauge.BarColor = ui.ColorRed
	a.gauge.BorderFg = ui.ColorWhite
	a.gauge.BorderLabelFg = ui.ColorCyan

	par1 := ui.NewPar("Up:Previous song|Down:Next song|Left:Vol Plus|Right:Vol Minus|Enter:Exit|q:Exit UI")
	par1.Height = 3
	par1.Width = 20
	par1.Y = 1
	par1.BorderLabel = "Player Options:"

	ui.Body.AddRows(
		ui.NewRow(
			ui.NewCol(6, 0, ls),
			ui.NewCol(6, 0, a.par0, a.gauge),
			//ui.NewCol(2, 0, g0),
		),
		ui.NewRow(
			ui.NewCol(12, 0, par1),
		),
	)

	ui.Body.Align()
	ui.Render(ui.Body)
	go func() {
		ui.Handle("q", func(ui.Event) {
			ui.StopLoop()
		})

		ui.Loop()
	}()

}

func (a *audios) updateUi() {
	a.par0.Text = a.audioList[a.playing]
	a.gauge.Percent = a.vol
	ui.Render(ui.Body)
}

func (a *audios) joystick() {

	echan := make(chan joyst.Event, 10)

	joy := joyst.Joystick{FilePath: "/dev/input/event0"}
	go joy.Poll(echan)
	for {
		code := <-echan
		switch code.Code {
		case joyst.LEFT:
			if a.vol != 0 {

				a.vol = a.vol - 10
				SetVol(a.vol)
				a.updateUi()
			}
			glog.Info("Detected: LEFT")

		case joyst.RIGHT:
			if a.vol != 100 {
				a.vol = a.vol + 10
				SetVol(a.vol)
				a.updateUi()
			}

			glog.Info("Detected: RIGHT")

		case joyst.UP:
			a.playing = (a.playing + 1) % len(a.audioList)
			a.cursor <- a.playing
			glog.Info("Detected: UP")

		case joyst.DOWN:
			if a.playing == 0 {
				a.playing = len(a.audioList) - 1
			} else {
				a.playing = a.playing - 1
			}
			a.cursor <- a.playing
			glog.Info("Detected: DOWN")

		case joyst.ENTER:
			glog.Info("Detected: ENTER")
			ui.StopLoop()
			ui.Clear()
			a.Stop <- 1

		default:
			glog.Errorf("Unknown code.Code : %d ", code.Code)
		}
		glog.Flush()
	}
}

func main() {
	flag.Parse()
	if soundpath == "" {
		glog.Fatal("Please Provide the path to the songs")
	}
	lastchr := soundpath[len(soundpath)-1:]
	if lastchr != "/" {
		soundpath = strings.Join([]string{soundpath, "/"}, "")
	}

	soundpath = strings.Join([]string{soundpath, "*.mp3"}, "")
	glog.Infof("looking for songs in %s", soundpath)
	a := new(audios)
	a.Init()
	a.displayInit()
	//echan := make(chan int, 10)

	//go playSound("/home/pi/InfraHack/player/sample1.mp3")
	go a.playSound()

	// Dipslay UI
	//go a.diplayUi()

	go a.joystick()
	<-a.Stop
	glog.Info("Closing the Playing...")
}
