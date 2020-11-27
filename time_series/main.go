package main

import (
	"bytes"
	"github.com/wcharczuk/go-chart"
	"io/ioutil"
	"math"
	"os"
)

func main() {

	maxX := 1000

	xValues := make([]float64, 0)
	yValues := make([]float64, 0)
	for i := maxX * -1; i < (maxX + 1); i++ {
		xValues = append(xValues, float64(i))
		yValues = append(yValues, fun(i))
	}

	xValues1 := make([]float64, 0)
	yValues1 := make([]float64, 0)
	for i := maxX * -1; i < (maxX + 1); i++ {
		xValues1 = append(xValues1, float64(i))
		yValues1 = append(yValues1, fun1(i))
	}

	graph := chart.Chart{
		Title: "Func Test",
		TitleStyle: chart.Style{
			Show: true,
		},
		ColorPalette: nil,
		Width:        0,
		Height:       0,
		Background: chart.Style{
			Show: true,
		},
		Canvas: chart.Style{
			Show: true,
		},
		XAxis: chart.XAxis{
			Name: "",
		},
		YAxis:          chart.YAxis{},
		YAxisSecondary: chart.YAxis{},
		Font:           nil,
		Series: []chart.Series{
			chart.ContinuousSeries{
				XValues: xValues,
				YValues: yValues,
			},
			chart.ContinuousSeries{
				XValues: xValues1,
				YValues: yValues1,
			},
		},
		Elements: nil,
	}

	buffer := bytes.NewBuffer([]byte{})
	err := graph.Render(chart.PNG, buffer)
	if err != nil {
		panic(err)
	}

	if err := ioutil.WriteFile("/Users/samuelorozco/dev_tools/time_series/test.png", buffer.Bytes(), os.ModePerm); err != nil {
		panic(err)
	}
}

func fun(x int) float64 {
	return math.Pow(float64(x*-3), 3) + 400
}
func fun1(x int) float64 {
	return math.Pow(float64(x*3), 2) + 800
}
