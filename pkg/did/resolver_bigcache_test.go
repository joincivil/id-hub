package did_test

import (
	"testing"
	"time"

	"github.com/allegro/bigcache"
	"github.com/joincivil/id-hub/pkg/did"
	"github.com/joincivil/id-hub/pkg/linkeddata"
	"github.com/joincivil/id-hub/pkg/utils"
	didlib "github.com/ockam-network/did"
)

var (
	testBigCacheConfig = bigcache.Config{
		Shards:             1024,
		LifeWindow:         2 * time.Second,
		CleanWindow:        3 * time.Second,
		MaxEntriesInWindow: 1000 * 10 * 60,
		MaxEntrySize:       500,
		Verbose:            true,
		HardMaxCacheSize:   16384,
	}
)

func TestBigCacheResolverCache(t *testing.T) {
	cache, _ := bigcache.NewBigCache(testBigCacheConfig)
	rcache := did.NewBigCacheResolverCache(cache)

	d, _ := didlib.Parse("did:web:uport.me")

	doc, err := rcache.Get(d)
	if err == nil {
		t.Errorf("Should have gotten an error for no result")
	}
	if doc != nil {
		t.Errorf("Should have gotten a nil doc")
	}
	if err != did.ErrResolverCacheNotFound {
		t.Errorf("Should have been an ErrResolverNotFound")
	}

	pk1 := did.CopyDID(d)
	owner := did.CopyDID(d)
	pk1.Fragment = "owner"

	doc = &did.Document{
		Context: "https://w3id.org/did/v1",
		ID:      *d,
		PublicKeys: []did.DocPublicKey{
			{
				ID:           pk1,
				Type:         linkeddata.SuiteTypeSecp256k1Verification2018,
				Owner:        owner,
				PublicKeyHex: utils.StrToPtr("042b0af9b3ae6c7c3a90b01a3879d9518081bc0dcdf038488db9cb109b082a77d97ea3373e3dfde0eccd9adbdce11d0302ea5c098dbb0b310234c86895c8641622"),
			},
		},
	}

	err = rcache.Set(d, doc)
	if err != nil {
		t.Errorf("Should have not gotten an error: err: %v", err)
	}

	doc, err = rcache.Get(d)
	if err != nil {
		t.Errorf("Should have not gotten an error: err: %v", err)
	}
	if doc == nil {
		t.Errorf("Should have gotten a doc")
	}

	time.Sleep(5 * time.Second)

	// Should have expired the entry and not get anything
	doc, err = rcache.Get(d)
	if err == nil {
		t.Errorf("Should have gotten an error for no result")
	}
	if doc != nil {
		t.Errorf("Should have gotten a nil doc")
	}
	if err != did.ErrResolverCacheNotFound {
		t.Errorf("Should have been an ErrResolverNotFound")
	}

}
