package common

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

const (
	BINARY = "BINARY(16)"
)

type BinUUID datatypes.BinUUID

func (u BinUUID) MarshalJSON() ([]byte, error) {
	return json.Marshal(datatypes.BinUUID(u).String())
}

func (u *BinUUID) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	if str == "" {
		*u = BinUUID{}
		return nil
	}

	uuid := datatypes.BinUUIDFromString(str)

	*u = BinUUID(uuid)
	return nil
}

// func (b BinUUID) ToBinUUID() datatypes.BinUUID {
// 	return datatypes.BinUUID(b)
// }
//
// // NewBinUUID 从 datatypes.BinUUID 创建 BinUUID
// func NewBinUUID(uuid datatypes.BinUUID) BinUUID {
// 	return BinUUID(uuid)
// }
//
// // NewBinUUIDFromString 从字符串创建 BinUUID
// func NewBinUUIDFromString(str string) (BinUUID, error) {
// 	uuid := datatypes.BinUUIDFromString(str)
// 	return BinUUID(uuid), nil
// }

func (BinUUID) GormDataType() string {
	return BINARY
}

// GormDBDataType gorm db data type.
func (BinUUID) GormDBDataType(db *gorm.DB, _ *schema.Field) string {
	switch db.Name() {
	case "mysql":
		return BINARY
	case "postgres":
		return "BYTEA"
	case "sqlserver":
		return BINARY
	case "sqlite":
		return "BLOB"
	default:
		return ""
	}
}

// Scan is the scanner function for this datatype.
func (u *BinUUID) Scan(value interface{}) error {
	valueBytes, ok := value.([]byte)
	if !ok {
		return errors.New("unable to convert value to bytes")
	}
	valueUUID, err := uuid.FromBytes(valueBytes)
	if err != nil {
		return err
	}
	*u = BinUUID(valueUUID)
	return nil
}

// Value is the valuer function for this datatype.
func (u BinUUID) Value() (driver.Value, error) {
	return uuid.UUID(u).MarshalBinary()
}

// String returns the string form of the UUID.
func (u BinUUID) Bytes() []byte {
	bytes, err := uuid.UUID(u).MarshalBinary()
	if err != nil {
		return nil
	}
	return bytes
}

// String returns the string form of the UUID.
func (u BinUUID) String() string {
	return uuid.UUID(u).String()
}

// Equals returns true if bytes form of BinUUID matches other, false otherwise.
func (u BinUUID) Equals(other BinUUID) bool {
	return bytes.Equal(u.Bytes(), other.Bytes())
}

// Length returns the number of characters in string form of UUID.
func (u BinUUID) LengthBytes() int {
	return len(u.Bytes())
}

// Length returns the number of characters in string form of UUID.
func (u BinUUID) Length() int {
	return len(u.String())
}

// IsNil returns true if the BinUUID is nil uuid (all zeroes), false otherwise.
func (u BinUUID) IsNil() bool {
	return uuid.UUID(u) == uuid.Nil
}

// IsEmpty returns true if BinUUID is nil uuid or of zero length, false otherwise.
func (u BinUUID) IsEmpty() bool {
	return u.IsNil() || u.Length() == 0
}

// IsNilPtr returns true if caller BinUUID ptr is nil, false otherwise.
func (u *BinUUID) IsNilPtr() bool {
	return u == nil
}

// IsEmptyPtr returns true if caller BinUUID ptr is nil or it's value is empty.
func (u *BinUUID) IsEmptyPtr() bool {
	return u.IsNilPtr() || u.IsEmpty()
}
