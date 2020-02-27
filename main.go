//go:generate enumer -type=Dimension -transform=snake -trimprefix=Dim -json -text
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/jessevdk/go-flags"
)

const (
	JSFormat = "var LambdaPlayers = %s;"
)

type Dimension int

const (
	DimOverworld Dimension = iota
	DimTheNether
	DimTheEnd
)

type options struct {
	CarpetURL         string `short:"c" long:"carpetURL"`
	CarpetSecret      string `short:"s" long:"carpetSecret"`
	Interval          int    `short:"i" long:"interval"`
	PortraitsFolder   string `short:"p" long:"portraits"`
	RootUnminedFolder string `short:"u" long:"unmined"`
}

type playerStatus struct {
	Name      string
	UUID      string
	X         float64
	Z         float64
	Dimension string
	Health    float32
	IsBot     bool
}

type serverStatus struct {
	Online    int
	StartTime int64
	Players   []*playerStatus
}

type jsPlayer struct {
	Name string  `json:"name"`
	X    float64 `json:"x"`
	Z    float64 `json:"z"`
	Bot  bool    `json:"bot"`
}

func main() {
	options := &options{
		Interval: 3600,
	}
	_, err := flags.Parse(options)

	if err != nil {
		panic(err)
	}

	pData := newPlayersData(options)

	sleepInterval := time.Duration(options.Interval) * time.Second
	client := &http.Client{}
	request, err := http.NewRequest("GET", options.CarpetURL, nil)
	if err != nil {
		panic(err)
	}

	request.Header.Set("X-Carpet", options.CarpetSecret)
	var status serverStatus

	for {
		resp, err := client.Do(request)
		if err != nil {
			fmt.Printf("Failed to call server: %s\n", err.Error())
			time.Sleep(sleepInterval)
			continue
		}

		err = json.NewDecoder(resp.Body).Decode(&status)
		if err != nil {
			fmt.Printf("Failed to decode json: %s\n", err.Error())
			time.Sleep(sleepInterval)
			continue
		}

		processPlayers(status, pData, options.RootUnminedFolder)

		time.Sleep(sleepInterval)
	}
}

func processPlayers(status serverStatus, data *playersData, unminedFolder string) {
	playersOW := make([]*jsPlayer, 0)
	playersTN := make([]*jsPlayer, 0)
	playersTE := make([]*jsPlayer, 0)

	for _, p := range status.Players {

		if !data.IsPortraitKnown(p.Name) {
			err := data.GetPortrait(p.UUID, p.Name)

			if err != nil {
				fmt.Printf("Failed to get player's [%s] portrait: %s\n", p.Name, err.Error())
				continue
			}
		}

		jp := &jsPlayer{
			Name: p.Name,
			X:    p.X,
			Z:    p.Z,
			Bot:  p.IsBot,
		}

		if strings.Contains(p.Dimension, ":") {
			p.Dimension = strings.Split(p.Dimension, "minecraft:")[1]
		}

		d, err := DimensionString(p.Dimension)
		if err != nil {
			fmt.Printf("Failed to get player's [%s] dimension [%s]: %s\n", p.Name, p.Dimension, err.Error())
			continue
		}

		switch d {
		case DimOverworld:
			playersOW = append(playersOW, jp)
			break
		case DimTheNether:
			playersTN = append(playersTN, jp)
			break
		default:
			playersTE = append(playersTE, jp)
		}
	}

	writeMarkers(playersOW, DimOverworld, unminedFolder)
	writeMarkers(playersTN, DimTheNether, unminedFolder)
	writeMarkers(playersTE, DimTheEnd, unminedFolder)
}

func writeMarkers(players []*jsPlayer, dim Dimension, unminedFolder string) {
	b, err := json.Marshal(players)

	if err != nil {
		fmt.Printf("Failed to marshal markers for dimension [%s]: %s\n", dim.String(), err.Error())
		return
	}

	r := fmt.Sprintf(JSFormat, string(b))
	err = ioutil.WriteFile(path.Join(unminedFolder, dim.String(), "lambda.players.js"), []byte(r), 0644)
	if err != nil {
		fmt.Printf("Failed to write markers for dimension [%s]: %s\n", dim.String(), err.Error())
	}
}
