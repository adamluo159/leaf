package jsonexcel

import (
	"fmt"
	"testing"
)

type AutoGenerated struct {
	Name        string `json:"name" file:"a"`
	URL         string `json:"url"`
	Page        int    `json:"page"`
	IsNonProfit bool   `json:"isNonProfit"`
	Address     struct {
		Street  string `json:"street"`
		City    string `json:"city"`
		Country string `json:"country"`
	} `json:"address"`
	Links []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"links"`
}

func TestReadJsonDir(t *testing.T) {
	Dir = "test"
	vmap := make(map[string]*AutoGenerated)
	Register(vmap)
	Init()

	for k, v := range vmap {
		fmt.Printf("%+v %+v\n", k, v)
	}
}