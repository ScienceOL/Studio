package model

import (
	"encoding/json"

	"github.com/scienceol/studio/service/pkg/utils"
)

type Position struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
	Z float32 `json:"z"`
}

// JSON 序列化
func (u Position) MarshalJSON() ([]byte, error) {
	type Alias Position
	return json.Marshal(Alias{
		X: utils.Or(u.X, 0),
		Y: utils.Or(u.Y, 0),
		Z: utils.Or(u.Z, 0),
	})
}

// JSON 反序列化
func (u *Position) UnmarshalJSON(data []byte) error {
	type Alias Position
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(u),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	u.X = utils.Or(u.X, 0)
	u.Y = utils.Or(u.Y, 0)
	u.Z = utils.Or(u.Z, 0)
	return nil
}

type Size struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// JSON 序列化
func (u Size) MarshalJSON() ([]byte, error) {
	type Alias Size
	return json.Marshal(Alias{
		Width:  utils.Or(u.Width, 200),
		Height: utils.Or(u.Height, 200),
	})
}

// JSON 反序列化
func (u *Size) UnmarshalJSON(data []byte) error {
	type Alias Size
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(u),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	u.Width = utils.Or(u.Width, 200)
	u.Height = utils.Or(u.Height, 200)
	return nil
}

type Scale struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
	Z float32 `json:"z"`
}

// JSON 序列化
func (u Scale) MarshalJSON() ([]byte, error) {
	type Alias Scale
	return json.Marshal(Alias{
		X: utils.Or(u.X, 1),
		Y: utils.Or(u.Y, 1),
		Z: utils.Or(u.Z, 1),
	})
}

// JSON 反序列化
func (u *Scale) UnmarshalJSON(data []byte) error {
	type Alias Scale
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(u),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	u.X = utils.Or(u.X, 1)
	u.Y = utils.Or(u.Y, 1)
	u.Z = utils.Or(u.Z, 1)
	return nil
}

type Rotation struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
	Z float32 `json:"z"`
}

// JSON 序列化
func (u Rotation) MarshalJSON() ([]byte, error) {
	type Alias Rotation
	return json.Marshal(Alias{
		X: utils.Or(u.X, 0),
		Y: utils.Or(u.Y, 0),
		Z: utils.Or(u.Z, 0),
	})
}

// JSON 反序列化
func (u *Rotation) UnmarshalJSON(data []byte) error {
	type Alias Rotation
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(u),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	u.X = utils.Or(u.X, 0)
	u.Y = utils.Or(u.Y, 0)
	u.Z = utils.Or(u.Z, 0)
	return nil
}

type Pose struct {
	Layout    string   `json:"layout"`
	Position  Position `json:"position"`
	Size      Size     `json:"size"`
	Scale     Scale    `json:"scale"`
	Rotation  Rotation `json:"rotation"`
	Disabled  bool     `json:"disabled"`
	Minimized bool     `json:"minimized"`
}
