package application

import (
	"context"
	"errors"
	"net/http"
	"testing"

	resellerdomain "github.com/dujiao-next/internal/modules/reseller/domain"

	"github.com/dujiao-next/internal/config"
)

type resellerResolverRepoStub struct {
	domain  *resellerdomain.Domain
	profile *resellerdomain.Profile
	err     error
	calls   int
}

func (s *resellerResolverRepoStub) FindActiveVerifiedDomain(host string) (*resellerdomain.Domain, error) {
	s.calls++
	if s.err != nil {
		return nil, s.err
	}
	return s.domain, nil
}

func (s *resellerResolverRepoStub) FindActiveProfileByCode(code string) (*resellerdomain.Profile, error) {
	s.calls++
	if s.err != nil {
		return nil, s.err
	}
	return s.profile, nil
}

func TestResellerDomainResolverDisabledTreatsEmptyHostAsMain(t *testing.T) {
	repo := &resellerResolverRepoStub{}
	resolver := NewDomainResolver(repo, config.ResellerConfig{Enabled: false})
	tenant, err := resolver.ResolveHost(context.Background(), "")
	if err != nil {
		t.Fatalf("disabled resolver should not fail on empty host: %v", err)
	}
	if !tenant.IsMain || tenant.Unavailable || tenant.ResellerID != nil {
		t.Fatalf("disabled resolver should return main tenant for empty host, got %+v", tenant)
	}
	if repo.calls != 0 {
		t.Fatalf("disabled resolver should not query repo, calls=%d", repo.calls)
	}
}

func TestResellerDomainResolverMainHost(t *testing.T) {
	repo := &resellerResolverRepoStub{}
	resolver := NewDomainResolver(repo, config.ResellerConfig{Enabled: true, MainHosts: []string{"main.example.test"}})
	tenant, err := resolver.ResolveHost(context.Background(), "MAIN.example.test:443")
	if err != nil {
		t.Fatalf("resolve main host failed: %v", err)
	}
	if !tenant.IsMain || tenant.Unavailable {
		t.Fatalf("expected main tenant, got %+v", tenant)
	}
	if repo.calls != 0 {
		t.Fatalf("main host should not query repo, calls=%d", repo.calls)
	}
}

func TestResellerDomainResolverActiveDomain(t *testing.T) {
	id := uint(7)
	repo := &resellerResolverRepoStub{domain: &resellerdomain.Domain{
		ID:         11,
		ResellerID: id,
		Domain:     "shop.example.test",
		Profile:    &resellerdomain.Profile{ID: id, UserID: 88, Status: resellerdomain.ProfileStatusActive},
	}}
	resolver := NewDomainResolver(repo, config.ResellerConfig{Enabled: true, MainHosts: []string{"main.example.test"}})
	tenant, err := resolver.ResolveHost(context.Background(), "shop.example.test")
	if err != nil {
		t.Fatalf("resolve reseller host failed: %v", err)
	}
	if tenant.IsMain || tenant.ResellerID == nil || *tenant.ResellerID != id {
		t.Fatalf("expected reseller tenant, got %+v", tenant)
	}
	if tenant.ResellerUserID != 88 {
		t.Fatalf("expected reseller user id 88, got %d", tenant.ResellerUserID)
	}
}

func TestResellerDomainResolverInactiveProfileUnavailable(t *testing.T) {
	id := uint(7)
	repo := &resellerResolverRepoStub{domain: &resellerdomain.Domain{
		ID:         11,
		ResellerID: id,
		Domain:     "shop.example.test",
		Status:     resellerdomain.DomainStatusActive,
		Profile:    &resellerdomain.Profile{ID: id, UserID: 88, Status: resellerdomain.ProfileStatusDisabled},
	}}
	resolver := NewDomainResolver(repo, config.ResellerConfig{Enabled: true, MainHosts: []string{"main.example.test"}})
	tenant, err := resolver.ResolveHost(context.Background(), "shop.example.test")
	if err != nil {
		t.Fatalf("resolve disabled profile host failed: %v", err)
	}
	if !tenant.Unavailable || tenant.ResellerID != nil {
		t.Fatalf("expected disabled profile to be unavailable, got %+v", tenant)
	}
}

func TestResellerDomainResolverRequestUsesTrustedForwardedHost(t *testing.T) {
	id := uint(8)
	repo := &resellerResolverRepoStub{domain: &resellerdomain.Domain{
		ID:         12,
		ResellerID: id,
		Domain:     "shop.example.test",
		Profile:    &resellerdomain.Profile{ID: id, UserID: 89, Status: resellerdomain.ProfileStatusActive},
	}}
	resolver := NewDomainResolver(repo, config.ResellerConfig{
		Enabled:              true,
		MainHosts:            []string{"main.example.test"},
		TrustedForwardedHost: true,
	})
	req, err := http.NewRequest(http.MethodGet, "https://internal.example.test/api/v1/public/config", nil)
	if err != nil {
		t.Fatalf("new request failed: %v", err)
	}
	req.Host = "internal.example.test"
	req.Header.Set("X-Forwarded-Host", "shop.example.test")
	tenant, err := resolver.ResolveRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("resolve request failed: %v", err)
	}
	if tenant.ResellerID == nil || *tenant.ResellerID != id {
		t.Fatalf("expected forwarded reseller tenant, got %+v", tenant)
	}
	if repo.calls != 1 {
		t.Fatalf("expected one repo lookup for forwarded host, calls=%d", repo.calls)
	}
}

func TestResellerDomainResolverUnknownDomainUnavailable(t *testing.T) {
	repo := &resellerResolverRepoStub{}
	resolver := NewDomainResolver(repo, config.ResellerConfig{Enabled: true, MainHosts: []string{"main.example.test"}})
	tenant, err := resolver.ResolveHost(context.Background(), "unknown.example.test")
	if err != nil {
		t.Fatalf("unknown domain should not return technical error: %v", err)
	}
	if !tenant.Unavailable || tenant.UnavailableReason != DomainUnavailableNotFound {
		t.Fatalf("expected unavailable not_found, got %+v", tenant)
	}
}

func TestResellerDomainResolverRepoError(t *testing.T) {
	repo := &resellerResolverRepoStub{err: errors.New("db down")}
	resolver := NewDomainResolver(repo, config.ResellerConfig{Enabled: true, MainHosts: []string{"main.example.test"}})
	_, err := resolver.ResolveHost(context.Background(), "shop.example.test")
	if err == nil {
		t.Fatal("expected repository error")
	}
}

func TestResellerDomainResolverAgentCodeOnMainHost(t *testing.T) {
	profileID := uint(99)
	userID := uint(1001)
	repo := &resellerResolverRepoStub{
		profile: &resellerdomain.Profile{
			ID:        profileID,
			UserID:    userID,
			Status:    resellerdomain.ProfileStatusActive,
			AgentCode: "AGENT99",
		},
	}
	resolver := NewDomainResolver(repo, config.ResellerConfig{Enabled: true, MainHosts: []string{"main.example.test"}})

	// 1. Query parameter ?agent=AGENT99
	req1, _ := http.NewRequest(http.MethodGet, "http://main.example.test/products?agent=AGENT99", nil)
	tenant1, err := resolver.ResolveRequest(context.Background(), req1)
	if err != nil {
		t.Fatalf("resolve request failed: %v", err)
	}
	if !tenant1.IsReseller() || tenant1.ResellerID == nil || *tenant1.ResellerID != profileID || tenant1.ResellerUserID != userID {
		t.Fatalf("expected reseller tenant from ?agent=AGENT99, got %+v", tenant1)
	}

	// 2. Header X-Agent-Code
	req2, _ := http.NewRequest(http.MethodGet, "http://main.example.test/api/v1/public/products", nil)
	req2.Header.Set("X-Agent-Code", "AGENT99")
	tenant2, err := resolver.ResolveRequest(context.Background(), req2)
	if err != nil {
		t.Fatalf("resolve request failed: %v", err)
	}
	if !tenant2.IsReseller() || *tenant2.ResellerID != profileID {
		t.Fatalf("expected reseller tenant from X-Agent-Code, got %+v", tenant2)
	}

	// 3. User reseller console path should bypass reseller tenant override
	req3, _ := http.NewRequest(http.MethodGet, "http://main.example.test/api/v1/user/reseller/profile?agent=AGENT99", nil)
	tenant3, err := resolver.ResolveRequest(context.Background(), req3)
	if err != nil {
		t.Fatalf("resolve request failed: %v", err)
	}
	if tenant3.IsReseller() || !tenant3.IsMain {
		t.Fatalf("expected main tenant for console endpoint, got %+v", tenant3)
	}
}

func TestResellerDomainResolverIgnoresLegacyAgentCookiesOnMainHost(t *testing.T) {
	repo := &resellerResolverRepoStub{
		profile: &resellerdomain.Profile{
			ID:        99,
			UserID:    1001,
			Status:    resellerdomain.ProfileStatusActive,
			AgentCode: "AGENT99",
		},
	}
	resolver := NewDomainResolver(repo, config.ResellerConfig{Enabled: true, MainHosts: []string{"main.example.test"}})
	req, err := http.NewRequest(http.MethodGet, "https://main.example.test/api/v1/public/products", nil)
	if err != nil {
		t.Fatalf("new request failed: %v", err)
	}
	req.Header.Set("Cookie", "dj_agent=AGENT99; agent_code=AGENT99")

	tenant, err := resolver.ResolveRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("resolve request failed: %v", err)
	}
	if !tenant.IsMain || tenant.IsReseller() || tenant.ResellerID != nil {
		t.Fatalf("legacy shared agent cookies must not change main-site pricing, got %+v", tenant)
	}
	if repo.calls != 0 {
		t.Fatalf("legacy shared agent cookies must not trigger a reseller lookup, calls=%d", repo.calls)
	}
}
