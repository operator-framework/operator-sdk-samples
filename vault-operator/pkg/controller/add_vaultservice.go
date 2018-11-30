package controller

import (
	"github.com/operator-framework/operator-sdk-samples/vault-operator/pkg/controller/vaultservice"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, vaultservice.Add)
}
