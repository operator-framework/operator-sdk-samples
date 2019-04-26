package main

import (
	"context"
	"fmt"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// podValidator admits a pod iff a specific annotation exists.
type podValidator struct {
	client  client.Client
	decoder *admission.Decoder
}

func (v *podValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	pod := &corev1.Pod{}

	err := v.decoder.Decode(req, pod)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	key := "mutating-memecahced"
	anno, found := pod.Annotations[key]
	if !found {
		return admission.Denied(fmt.Sprintf("missing mutation %s", key))
	}
	if anno != "new-annotation" {
		return admission.Denied(fmt.Sprintf("annotation %s did not have mutation %q", key, "new-annotation"))
	}

	return admission.Allowed("")
}

// podValidator implements inject.Client.
// A client will be automatically injected.

// InjectClient injects the client.
func (v *podValidator) InjectClient(c client.Client) error {
	v.client = c
	return nil
}

// podValidator implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (v *podValidator) InjectDecoder(d *admission.Decoder) error {
	v.decoder = d
	return nil
}
