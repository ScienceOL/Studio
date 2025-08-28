package utils

import "testing"

func TestJwt(t *testing.T) {
	token := `eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTY2OTY4OTQsImlkZW50aXR5Ijp7ImVudiI6InByb2QiLCJvcmdJZCI6MTgwMjQsInVzZXJJZCI6MTgxNjl9LCJvcmlnX2lhdCI6MTc1NjA5MjA5NH0.QTOLztLYPK1bx1aFZ2s4I_4a3K3dvvG5e9Lv-R4Ch5MNcyKEqy-J6BGDSET5ODJpPOPs3PiL4hS0a7GQGhJxIAjZrs3F2SzSS-b5CC_X7goOWcf2xlBIji5BRgTFJooVzAcTkFrmOYJwwSuTfNxAwIZGJG5-2yGjZH5c9VobB9VfQLCXGsLQYM5IBEgPpSDdWXudnjL7hUzvOtBA03Z1m2K0kuQYVH42K5vhih_0KYFbhklQaK8tOHVPwuaW0uQMle-VmX8RnezkReMnZ9G6ul1venq7iP27xGGfUih-HFJk8eo37szJID9vIoV760CJ2CtoOd4297LWsNMd-io7mA`
	token1 := `eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTY5Njg3ODgsImlkZW50aXR5Ijp7ImVudiI6InRlc3QiLCJvcmdJZCI6MTE1OTg3LCJ1c2VySWQiOjExMzQ0Nn0sIm9yaWdfaWF0IjoxNzU2MzYzOTg4fQ.SH_0FIL52XqtIH8qwi4_sEaz2jbYcDKrS23sZAvRRGf2KBlX2RACsYqeQZl8WixdATQjaUer9EbtSbpVzKZBnaeGEqgcLmxQazhBu2QbK2pOcGjJdJLm0xpe5nPhMqQXUaR2jA9erkEtmOiz6btw_cA4_Z0MTbbeMGq3CzYGI39coy1zrUzOVRpPs6EtkxAA0nxYpuXKZ7a8YckyyrZkrQq-AXGCWn893BN-j0uSLc73GXqKKa3-ZdZ4bCD_Gy7Y-O3haqewx1n8ZKiuA7HFXw-uuqmut7I7ugIL8anY9Q3_cAQE_sPmnPefzurRNlYj6fcLGchu0DQITZ0GyzoJuw`

	_ = token
	data := &Claims{}
	err := ParseJWTWithPublicKey(token1, DefaultPublicKey, data)
	if err != nil {
		t.Errorf("parse jet err: %+v", err)
	}

	err = ParseJWTWithSecret(token1, "dp666666", data)
	if err != nil {
		t.Errorf("parse jet err: %+v", err)
	}
}
