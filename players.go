package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"image"
	"image/png"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
)

type playersData struct {
	portraitsFolder string
}

func newPlayersData(options *options) *playersData {
	return &playersData{
		portraitsFolder: options.PortraitsFolder,
	}
}

func (p *playersData) IsPortraitKnown(name string) bool {
	file := path.Join(p.portraitsFolder, name+".png")

	_, err := os.Stat(file)

	return err == nil
}

func (p *playersData) GetPortrait(uuid string, name string) error {
	uuid = strings.Replace(strings.ToLower(uuid), "-", "", -1)

	skin, err := getSkinImage(uuid)
	if err != nil {
		return err
	}

	if skin == nil {
		return p.copyAlex(name)
	}

	head := skin.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(relative(8, 8, 8, 8))

	f, err := os.Create(path.Join(p.portraitsFolder, name+".png"))
	if err != nil {
		return err
	}

	defer f.Close()
	return png.Encode(f, head)
}

func (p *playersData) copyAlex(name string) error {
	in, err := os.Open(path.Join(p.portraitsFolder, "Alex.png"))
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(path.Join(p.portraitsFolder, name+".png"))
	if err != nil {
		return err
	}

	_, err = io.Copy(out, in)

	if err != nil {
		return err
	}

	return out.Close()
}

func relative(x, y, width, height int) image.Rectangle {
	return image.Rect(x, y, x+width, y+height)
}

type mcResp struct {
	Properties []*mcProp `json:"properties"`
}

type mcProp struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type mcSkinURL struct {
	URL string `json:"url"`
}

type mcSkin struct {
	SkinUrl mcSkinURL `json:"SKIN"`
}

type mcSkinResponse struct {
	Textures mcSkin `json:"textures"`
}

func getSkinImage(uuid string) (image.Image, error) {
	resp, err := http.Get("https://sessionserver.mojang.com/session/minecraft/profile/" + uuid)
	if err != nil {
		return nil, errors.New("failed to call mojang API")
	}

	r := &mcResp{}

	err = json.NewDecoder(resp.Body).Decode(r)
	if err != nil {
		return nil, err
	}

	texture := ""

	for _, v := range r.Properties {
		if v.Name == "textures" {
			texture = v.Value
			break
		}
	}

	if "" == texture {
		return nil, errors.New("skin not found")
	}

	b64, err := base64.StdEncoding.DecodeString(texture)

	if err != nil {
		return nil, err
	}

	skin := &mcSkinResponse{}

	err = json.NewDecoder(bytes.NewReader(b64)).Decode(skin)
	if err != nil {
		return nil, err
	}

	if "" == skin.Textures.SkinUrl.URL {
		return nil, nil
	}

	resp, err = http.Get(skin.Textures.SkinUrl.URL)
	if err != nil {
		return nil, errors.New("failed to get skin")
	}

	src, _, err := image.Decode(resp.Body)

	return src, err
}
