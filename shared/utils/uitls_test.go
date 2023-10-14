package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type Data struct {
	DataField string `json:"dataField"`
}

func TestFromJson(t *testing.T) {
	t.Run("Successfully unmarshal data", func(t *testing.T) {
		testData := []byte(`{"dataField":"value"}`)
		expectedData := Data{
			DataField: "value",
		}
		actualData, err := FromJson[Data](testData)
		assert.Nil(t, err)
		assert.Equal(t, expectedData, actualData)
	})
	t.Run("If unmarshalling fails, error is returned", func(t *testing.T) {
		testData := []byte("wrongData")
		actualData, err := FromJson[Data](testData)
		assert.Error(t, err)
		assert.Empty(t, actualData)
	})
}

func TestToJson(t *testing.T) {
	t.Run("Successfully encode data to JSON", func(t *testing.T) {
		testData := Data{
			DataField: "value",
		}
		expectedData := []byte(`{"dataField":"value"}`)

		actualData, err := ToJson[Data](testData)
		assert.Nil(t, err)
		assert.Equal(t, expectedData, actualData)
	})
	t.Run("If marshalling fails, error is returned", func(t *testing.T) {
		type wrongType func()
		actualData, err := ToJson[wrongType](func() {})
		assert.Error(t, err)
		assert.Empty(t, actualData)
	})
}
