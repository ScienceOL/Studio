package model

import "encoding/json"

type BinUUID datatypes.BinUUID

func (b BinUUID) MarshalJSON() ([]byte, error) {
	return json.Marshal(datatypes.BinUUID(b).String())
}

func (b *BinUUID) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	if str == "" {
		*b = BinUUID{}
		return nil
	}

	uuid, err := datatypes.BinUUID{}.FromString(str)
	if err != nil {
		return err
	}

	*b = BinUUID(uuid)
	return nil
}

func (b BinUUID) String() string {
	return datatypes.BinUUID(b).String()
}

func (b BinUUID) ToBinUUID() datatypes.BinUUID {
	return datatypes.BinUUID(b)
}

// NewBinUUID 从 datatypes.BinUUID 创建 BinUUID
func NewBinUUID(uuid datatypes.BinUUID) BinUUID {
	return BinUUID(uuid)
}

// NewBinUUIDFromString 从字符串创建 BinUUID
func NewBinUUIDFromString(str string) (BinUUID, error) {
	uuid, err := datatypes.BinUUID{}.FromString(str)
	if err != nil {
		return BinUUID{}, err
	}
	return BinUUID(uuid), nil
}
