package handler

import (
	"github.com/stretchr/testify/assert"
	"go-es/internal/service/index/model"
	"testing"
)

func TestGenerateProperties(t *testing.T) {
	fields := map[string]model.FieldConfig{
		"name": {
			Type:         "text",
			Autocomplete: true,
			Search:       true,
		},
		"age": {
			Type: "integer",
		},
	}

	properties, err := generateProperties(fields)
	assert.NoError(t, err)
	assert.Equal(t, "text", properties["name"].(map[string]interface{})["type"])
	assert.Equal(t, "integer", properties["age"].(map[string]interface{})["type"])
}
