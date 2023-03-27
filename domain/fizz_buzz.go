package domain

import "encoding/json"

type FizzBuzzRequest struct {
	FstModulo int    `json:"fst_mod" binding:"required,gte=1"`
	SndModulo int    `json:"snd_mod" binding:"required,gte=1"`
	Limit     int    `json:"limit" binding:"required,gte=1"`
	FstStr    string `json:"fst_str" binding:"required"`
	SndStr    string `json:"snd_str" binding:"required"`
}

func (fbr *FizzBuzzRequest) ToBytes() []byte {
	data, err := json.Marshal(fbr)
	if err != nil {
		return []byte{}
	}
	return data
}

func FromStrToRequestFB(payload string) *FizzBuzzRequest {
	var fbr FizzBuzzRequest
	err := json.Unmarshal([]byte(payload), &fbr)
	if err != nil {
		return nil
	}
	return &fbr
}
