package helper

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/thalalhassan/edu_management/internal/shared/response"
)

func PrintStruct(s any, message string) {
	fmt.Println("=======>", message)
	data, _ := json.MarshalIndent(s, "", "  ")
	fmt.Println(string(data))
	fmt.Println("<=======")
}

func ParseParamUUIDWithAbort(c *gin.Context, param string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(param))
	if err != nil {
		response.BadRequest(c, "invalid ID format")
		// c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid UUID for param: %s", param)})
		return uuid.Nil, false
	}
	return id, true
}

func ParseUUID(idStr string) (uuid.UUID, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil

}
