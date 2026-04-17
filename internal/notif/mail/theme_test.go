package mail

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestName(t *testing.T) {
	assert.Equal(t, "ftpgrab", new(Theme).Name())
}

func TestTemplates(t *testing.T) {
	theme := new(Theme)

	assert.Contains(t, theme.HTMLTemplate(), `status_{{ $cell.Value }}.png`)
	assert.Contains(t, theme.HTMLTemplate(), `{{.Hermes.Product.Copyright}}`)
	assert.Contains(t, theme.PlainTextTemplate(), `{{ $action.Button.Link }}`)
	assert.Contains(t, theme.PlainTextTemplate(), `{{ $cell.Value }}`)
}
