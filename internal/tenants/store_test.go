package tenants

import (
	"testing"

	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/types"
)

func TestStoreCRUD(t *testing.T) {
	s := NewStore()
	s.EnsureDefault("tnt_default")

	if _, ok := s.Get("tnt_default"); !ok {
		t.Fatalf("expected default tenant to exist")
	}

	err := s.Create(types.Tenant{TenantID: "tnt_new", Name: "New"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := s.Create(types.Tenant{TenantID: "tnt_new"}); err == nil {
		t.Fatalf("expected duplicate create to fail")
	}

	updated, err := s.Update("tnt_new", types.Tenant{Status: "suspended"})
	if err != nil || updated.Status != "suspended" {
		t.Fatalf("update failed: %v", err)
	}

	if _, err := s.Delete("tnt_default"); err == nil {
		t.Fatalf("expected delete default to fail")
	}
	if _, err := s.Delete("tnt_new"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}
