package stub

import (
	api "github.com/operator-framework/operator-sdk-samples/vault-operator/pkg/apis/vault/v1alpha1"
	"github.com/operator-framework/operator-sdk-samples/vault-operator/pkg/vault"

	"github.com/coreos/operator-sdk/pkg/sdk/handler"
	"github.com/coreos/operator-sdk/pkg/sdk/types"
)

func NewHandler() handler.Handler {
	return &Handler{}
}

type Handler struct {
	// Fill me
}

func (h *Handler) Handle(ctx types.Context, event types.Event) error {
	switch o := event.Object.(type) {
	case *api.VaultService:
		return vault.Reconcile(o)
	}
	return nil
}
