package tests

import (
	"github.com/focks/sanity"
	"testing"
)

type FlatStruct struct {
	Id          string  `json:"id" sanity:"notnull,maxlen" notnull:"true" maxlen:"5"`
	Description *string `json:"description" sanity:"notnull" notnull:"true"`
	UsedCount   int64   `json:"used_count" sanity:"notnull,gt" gt:"23" notnull:"true"`
}


func TestFlatStruct(t *testing.T) {
	description := "this is just a sample description"
	obj := FlatStruct{
		Id:          "",
		Description: &description,
		UsedCount:   24,
	}

	t.Run("flat struct", func(t *testing.T){
		if vinfo, ok := sanity.Check(obj); !ok {
			t.Error(vinfo.Errors)
		}

	})
}