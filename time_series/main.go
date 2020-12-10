package main

import (
	"bytes"
	"dev_tools/files"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

type Point struct {
	X float64
	Y float64
}

var (
	rootCmd = &cobra.Command{
		Use:   "ts",
		Short: "Time Series from csv",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 2 {
				panic("must provide data file and image name")
			}
			BuildTimeSeries(args[0], args[1])
		},
	}
)

var lineColors = []drawing.Color{
	{
		R: 255, G: 51, B: 51, A: 100,
	},
	{
		R: 255, G: 153, B: 51, A: 100,
	},
	{
		R: 255, G: 255, B: 102, A: 100,
	},
	{
		R: 178, G: 255, B: 102, A: 100,
	},
}

func main() {
	Execute()
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func BuildTimeSeries(dataFilePath string, fileName string) {
	// if multiple or single data series files
	fileNames := make([]string, 0)
	if strings.Contains(dataFilePath, ",") {
		fileNames = append(fileNames, strings.Split(dataFilePath, ",")...)
	} else {
		fileNames = append(fileNames, dataFilePath)
	}
	Build(fileNames, fileName)
}

func Build(fileNames []string, fileName string) {
	series := make([]chart.Series, 0)
	for i := range fileNames {
		current := fileNames[i]
		dataPoints := LoadDataPointsFromFile(current)
		seriesFromDataPoints := GetContinuousSeriesFromDataPoints(dataPoints, current, i)
		series = append(series, seriesFromDataPoints)
	}

	graph := chart.Chart{
		Title:        fileName,
		TitleStyle:   chart.Style{},
		ColorPalette: nil,
		Width:        0,
		Height:       0,
		Series:       series,
		XAxis: chart.XAxis{
			Name:           "",
			NameStyle:      chart.Style{},
			Style:          chart.Style{},
			ValueFormatter: nil,
			Range:          nil,
			TickStyle:      chart.Style{},
			Ticks:          nil,
			TickPosition:   0,
			GridLines:      nil,
			GridMajorStyle: chart.Style{},
			GridMinorStyle: chart.Style{},
		},
	}

	buffer := bytes.NewBuffer([]byte{})
	err := graph.Render(chart.PNG, buffer)
	if err != nil {
		panic(err)
	}

	if err := ioutil.WriteFile(fileName, buffer.Bytes(), os.ModePerm); err != nil {
		panic(err)
	}
}

func GetContinuousSeriesFromDataPoints(data []*Point, name string, idx int) chart.ContinuousSeries {
	xVAl := make([]float64, len(data))
	yVAl := make([]float64, len(data))
	for i := range data {
		xVAl[i] = data[i].X
		yVAl[i] = data[i].Y
	}

	return chart.ContinuousSeries{
		Style: chart.Style{
			StrokeWidth: 0,
			StrokeColor: getColorFromIdx(idx),
		},
		Name:    name,
		XValues: xVAl,
		YValues: yVAl,
	}
}

func LoadDataPointsFromFile(fileName string) []*Point {
	val, err := files.ReadStringFromFile(fileName)
	if err != nil {
		panic(err)
	}
	lines := strings.Split(val, "\n")
	result := make([]*Point, len(lines))
	for i := range lines {
		result[i] = GetPointFromLine(lines[i], i)
	}
	return result
}

func GetPointFromLine(line string, idx int) *Point {
	segs := strings.Split(line, ",")

	xVal, err := strconv.ParseFloat(segs[0], 64)
	if err != nil {
		panic(fmt.Sprintf("unable to parse x value on line %d", idx+1))
	}

	yVal, err := strconv.ParseFloat(segs[1], 64)
	if err != nil {
		panic(fmt.Sprintf("unable to parse y value on line %d", idx+1))
	}

	return &Point{
		X: xVal,
		Y: yVal,
	}
}

func getColorFromIdx(idx int) drawing.Color {
	if idx > len(lineColors)-1 {
		idx = 0
	}
	return lineColors[idx]
}
