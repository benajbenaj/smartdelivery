package main

import "testing"

func TestOpenAPIPathConvertsGinParams(t *testing.T) {
	got := openAPIPath("/shops/:shop/delivery-rules/:id")
	want := "/shops/{shop}/delivery-rules/{id}"

	if got != want {
		t.Fatalf("openAPIPath() = %q, want %q", got, want)
	}
}

func TestGenerateSpecIncludesRegisteredRoutes(t *testing.T) {
	spec, err := generateSpec()
	if err != nil {
		t.Fatalf("generateSpec() error = %v", err)
	}

	if _, ok := spec.Paths["/healthz"]["get"]; !ok {
		t.Fatalf("spec missing GET /healthz")
	}
	if _, ok := spec.Paths["/shops/{shop}/delivery-rules"]["post"]; !ok {
		t.Fatalf("spec missing POST /shops/{shop}/delivery-rules")
	}
	if _, ok := spec.Paths["/shops/{shop}/delivery-rules/imports"]["post"]; !ok {
		t.Fatalf("spec missing POST /shops/{shop}/delivery-rules/imports")
	}
	if _, ok := spec.Paths["/shopify/install"]["get"]; !ok {
		t.Fatalf("spec missing GET /shopify/install")
	}
	if _, ok := spec.Paths["/shopify/callback"]["get"]; !ok {
		t.Fatalf("spec missing GET /shopify/callback")
	}
}
