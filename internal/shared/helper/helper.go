package helper

import (
	"encoding/json"
	"fmt"
)

func PrintStruct(s any) {
	fmt.Println("=======>")
	data, _ := json.MarshalIndent(s, "", "  ")
	fmt.Println(string(data))
	fmt.Println("<=======")
}
