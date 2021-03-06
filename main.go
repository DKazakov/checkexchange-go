package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/nsf/termbox-go"
	"github.com/wcharczuk/go-chart"
	drawing "github.com/wcharczuk/go-chart/drawing"
	"gopkg.in/xmlpath.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var (
	arr                     []float64
	date                    []float64
	lastupdate, lastrequest string
)

func main() {
	f, _ := os.OpenFile("/var/log/self/checkexchange.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer f.Close()
	log.SetOutput(f)
	syscall.Dup2(int(f.Fd()), 2)

	termbox.Init()
	termbox.SetOutputMode(termbox.OutputMode(termbox.OutputNormal))
	defer termbox.Close()

	nextIteration(0, 0)

	ticker := time.NewTicker(1 * 60 * time.Second)

	go func() {
		for _ = range ticker.C {
			sizeX, sizeY := termbox.Size()
			nextIteration(sizeX*7, (sizeY-4)*15)
		}
	}()

loop:
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc:
				ticker.Stop()
				log.Println("stop timer and exit")
				break loop
			case 0:
				switch ev.Ch {
				case 113:
					ticker.Stop()
					break loop
				default:
					log.Printf("%+v", ev)
				}
			default:
				log.Printf("%+v", ev)
			}
		}
	}
}

func nextIteration(width, height int) {
	const (
		limit     = 1
		fixedpart = 1668190.2879
		val       = 1400.0
	)

	price, err := getPrice()
	now := time.Now()
	lastrequest = now.Format("15:04:05")
	if err == nil {
		if len(arr) > 0 && price != arr[len(arr)-1] || len(arr) == 0 {
			lastupdate = now.Format("15:04:05")
			arr = append(arr, price)
			date = append(date, float64(now.Unix()))
			fmt.Printf("\x1b[0;0H56.36 * 10000 + 56.53 * 10000 + 60.84 * 1000 + 60.49 * 3266.71 + 60.56 * 1000 + 61.19 * 3500 + 61.22 * 100 + %.2f * %.2f = %s ", price, val, formatNumber(fixedpart+price*val, " "))
			last := 0
			if len(arr) > limit {
				last = len(arr) - limit
			}
			lastprice := arr[0]
			for i, e := range arr {
				if i > last {
					if e > lastprice {
						fmt.Printf("\x1b[48;05;34m%s\x1b[0m", string(8593))
					} else {
						fmt.Printf("\x1b[48;05;196m%s\x1b[0m", string(8595))
					}
					lastprice = e
				}
			}
			fmt.Printf("\n%.2f * %.2f = %s\n", price, val, formatNumber(price*val, " "))
			if last > 0 {
				image := render(width, height)
				str := base64.StdEncoding.EncodeToString(image.Bytes())
				fmt.Printf("\x1b[%d;%dH\x1b]1337;File=name=none;size=%d;inline=1:%s\a", 5, 0, len(str), str)
			}
		}
	}
	fmt.Printf("\x1b[3;0Hlast request: %s, last update: %s", lastrequest, lastupdate)
}

func getPrice() (price float64, err error) {
	resp, err := http.Get("https://www.raiffeisen.ru/currency_rates/")
	if err != nil {
		log.Println("wrong request, error: ", err)
	} else {
		body, err := ioutil.ReadAll(resp.Body)
		if resp.StatusCode != 200 {
			log.Println("non-200 response")
		} else if err != nil {
			log.Println("Parse error: ", err)
		} else {
			reader := strings.NewReader(string(body))
			xmlroot, err := xmlpath.ParseHTML(reader)
			if err != nil {
				log.Println("xml parse error, wrong html")
			} else {
				path := xmlpath.MustCompile(`//*[@id="online"]/div[2]/div/div/div[2]/div[4]`)
				if value, ok := path.String(xmlroot); ok {
					price, _ = strconv.ParseFloat(value, 64)
				} else {
					err = errors.New("xml parse error, wrong xpath")
				}
			}
		}
	}

	return
}

func formatNumber(i float64, divider string) string {
	var out = ""
	for ; i >= 1000.0; i = i / 1000.0 {
		out = fmt.Sprintf("%s%03d", divider, int(i)%1000) + out
	}
	return fmt.Sprintf("%d%s", int(i), out)
}

func render(imageWidth, imageHeight int) *bytes.Buffer {
	var (
		min, max float64
	)

	buffer := bytes.NewBuffer([]byte{})
	series := []chart.Series{}

	min = arr[0]
	max = arr[0]
	for _, e := range arr {
		if max < e {
			max = e
		}
		if min > e {
			min = e
		}
	}

	series = append(series, chart.ContinuousSeries{
		Style: chart.Style{
			Show:        true,
			StrokeColor: drawing.Color{R: 255, G: 0, B: 0, A: 255},
			FillColor:   drawing.Color{R: 255, G: 0, B: 0, A: 255},
		},
		YValues: arr,
		XValues: date,
	})

	graph := chart.Chart{
		Width:  imageWidth,
		Height: imageHeight,
		Background: chart.Style{
			Padding: chart.Box{
				Top:    20,
				Left:   0,
				Right:  0,
				Bottom: 0,
			},
		},
		XAxis: chart.XAxis{
			Style: chart.Style{
				Show:     true,
				FontSize: 7.0,
			},
			TickPosition: chart.TickPositionBetweenTicks,
			ValueFormatter: func(v interface{}) string {
				typed := v.(float64)
				typedDate := time.Unix(0, int64(typed)*1000000000)
				return fmt.Sprintf("%.2d:%.2d", typedDate.Hour(), typedDate.Minute())
			},
		},
		YAxis: chart.YAxis{
			Style: chart.Style{
				Show:     true,
				FontSize: 7.0,
			},
			Range: &chart.ContinuousRange{
				Max: max + .1,
				Min: min - .1,
			},
		},
		YAxisSecondary: chart.YAxis{
			Style: chart.Style{
				Show:     true,
				FontSize: 7.0,
			},
			Range: &chart.ContinuousRange{
				Max: max + .1,
				Min: min - .1,
			},
		},
		Series: series,
	}

	graph.Render(chart.PNG, buffer)

	return buffer
}
