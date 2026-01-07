package schemas

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntoMarkdown(t *testing.T) {
	assert.NotEmpty(t, IntoMarkdown(Details()))
}
