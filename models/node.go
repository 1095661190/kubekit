package models

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type NodeList struct {
	Nodes []Node `json:"nodes"`
}

type Node struct {
	Name      string `json:"name"`
	IP        string `json:"ip"`
	CreatedAt string `json:"createdAt"`
	Status    int    `josn:"status"`
	Comment   string `json:"comment"`
}

func (n *NodeList) Serialize(p interface{}, file string) error {
	bytes, err := json.Marshal(p)

	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	ioutil.WriteFile("./.nodes", bytes, os.ModeAppend)
	return nil
}

func (n *NodeList) Deserialize() (*NodeList, error) {
	raw, err := ioutil.ReadFile("./.nodes")

	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	node := &NodeList{}
	json.Unmarshal(raw, node)
	return node, nil
}
