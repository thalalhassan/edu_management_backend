package helper

import (
	"encoding/json"
	"fmt"
)

func PrintStruct(s any, message string) {
	fmt.Println("=======>", message)
	data, _ := json.MarshalIndent(s, "", "  ")
	fmt.Println(string(data))
	fmt.Println("<=======")
}
