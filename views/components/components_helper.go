// Package components provides shared UI components.
package components

import (
	"boilerplate/internal/models"
	"context"

	"github.com/a-h/templ"
)

func navigation(ctx context.Context) templ.Component {
	if u := models.UserFromContext(ctx); u == nil {
		return unConnected()
	}
	return connected()
}
