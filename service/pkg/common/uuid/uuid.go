//nolint:revive // var-naming: common package contains shared utilities
package uuid

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/gofrs/uuid/v5"
)

type UUID struct {
	uuid.UUID
}

func FromString(value string) (UUID, error) {
	parsed, err := uuid.FromString(value)
	if err != nil {
		return UUID{}, err
	}
	return UUID{UUID: parsed}, nil
}

// // 添加这个方法来支持 URI 参数绑定
// func (u *UUID) UnmarshalText(text []byte) error {
// 	s := string(text)
// 	if s == "" {
// 		u.UUID = uuid.Nil
// 		return nil
// 	}
//
// 	parsed, err := uuid.FromString(s)
// 	if err != nil {
// 		return err
// 	}
// 	u.UUID = parsed
// 	return nil
// }
//
// // 可选：添加 MarshalText 方法保持对称性
// func (u UUID) MarshalText() ([]byte, error) {
// 	if u.UUID.IsNil() {
// 		return []byte(""), nil
// 	}
// 	return []byte(u.UUID.String()), nil
// }

// 实现 Gin 的 BindUnmarshaler 接口，用于 URI 参数绑定
func (u *UUID) UnmarshalParam(param string) error {
	if param == "" {
		u.UUID = uuid.Nil
		return nil
	}

	parsed, err := uuid.FromString(param)
	if err != nil {
		return err
	}
	u.UUID = parsed
	return nil
}

// JSON 序列化
func (u UUID) MarshalJSON() ([]byte, error) {
	if u.IsNil() {
		return json.Marshal("")
	}
	return json.Marshal(u.String())
}

// JSON 反序列化
func (u *UUID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	if s == "" {
		u.UUID = uuid.Nil
		return nil
	}

	parsed, err := uuid.FromString(s)
	if err != nil {
		return err
	}
	u.UUID = parsed
	return nil
}

// 实现 driver.Valuer 接口 - 用于数据库写入
func (u UUID) Value() (driver.Value, error) {
	if u.IsNil() {
		return nil, nil
	}
	return u.String(), nil
}

// 实现 sql.Scanner 接口 - 用于数据库读取
func (u *UUID) Scan(value any) error {
	if value == nil {
		u.UUID = uuid.Nil
		return nil
	}

	switch v := value.(type) {
	case string:
		if v == "" {
			u.UUID = uuid.Nil
			return nil
		}
		parsed, err := uuid.FromString(v)
		if err != nil {
			return err
		}
		u.UUID = parsed
		return nil
	case []byte:
		if len(v) == 0 {
			u.UUID = uuid.Nil
			return nil
		}
		parsed, err := uuid.FromString(string(v))
		if err != nil {
			return err
		}
		u.UUID = parsed
		return nil
	default:
		return fmt.Errorf("cannot scan %T into CustomUUID", value)
	}
}

// 实现 GORM Schema 接口（可选）
func (UUID) GormDataType() string {
	return "uuid"
}

func NewV4() UUID {
	return UUID{UUID: uuid.Must(uuid.NewV4())}
}

func NewNil() UUID {
	return UUID{UUID: uuid.Nil}
}

func (u UUID) String() string {
	return u.UUID.String()
}
