package schemas

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NoPanic(t *testing.T) {
	assert.NotNil(t, Details(), "Details must not panic and provide proper result")
}

func Test_RequiredFields(t *testing.T) {
	for _, detail := range Details() {
		assert.False(t, detail.Name == "" && detail.Tag == "", "Either Name or Tag must not be empty")
		label := fmt.Sprintf("detail name: %s", detail.Name)
		if detail.Name == "" {
			label = fmt.Sprintf("detail tag: %s", detail.Tag)
		}
		assert.NotEmptyf(t, detail.Description, "Description must not be empty for %s", label)
		assert.NotEmpty(t, detail.When, "When must not be empty for %s", label)
		assert.NotEmpty(t, detail.Examples, "Examples must not be empty for %s", label)
		assert.NotEmpty(t, detail.Value, "Value must not be empty for %s", label)
	}
}

func Test_UniqueNames(t *testing.T) {
	names := make(map[string]bool)
	for _, detail := range Details() {
		if detail.Name == "" {
			continue
		}
		if _, ok := names[detail.Name]; ok {
			t.Errorf("Duplicate name: %s", detail.Name)
		}
		names[detail.Name] = true
	}
}

func Test_UniqueTagsWithoutNames(t *testing.T) {
	tags := make(map[string]bool)
	for _, detail := range Details() {
		if detail.Tag == "" || detail.Name != "" {
			continue
		}
		if _, ok := tags[detail.Tag]; ok {
			t.Errorf("Duplicate tag without a name: %s", detail.Tag)
		}
		tags[detail.Tag] = true
	}
}

func Test_ValidSchemas(t *testing.T) {
	for _, ds := range Details() {
		assert.NotNil(t, ds.ValueSchema())
	}
}
