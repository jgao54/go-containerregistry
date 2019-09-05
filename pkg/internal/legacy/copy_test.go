package legacy

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/registry"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

func TestCopySchema1(t *testing.T) {
	// Set up a fake registry.
	s := httptest.NewServer(registry.New())
	defer s.Close()
	u, err := url.Parse(s.URL)
	if err != nil {
		t.Fatal(err)
	}

	// We'll copy from src to dst.
	src := fmt.Sprintf("%s/schema1/src", u.Host)
	srcRef, err := name.ParseReference(src)
	if err != nil {
		t.Fatal(err)
	}
	dst := fmt.Sprintf("%s/schema1/dst", u.Host)
	dstRef, err := name.ParseReference(dst)
	if err != nil {
		t.Fatal(err)
	}

	// Create a random layer.
	layer, err := random.Layer(1024, types.DockerLayer)
	if err != nil {
		t.Fatal(err)
	}
	digest, err := layer.Digest()
	if err != nil {
		t.Fatal(err)
	}
	layerRef, err := name.NewDigest(fmt.Sprintf("%s@%s", src, digest))
	if err != nil {
		t.Fatal(err)
	}

	// Populate the registry with a layer and a schema 1 manifest referencing it.
	if err := remote.WriteLayer(layerRef, layer); err != nil {
		t.Fatal(err)
	}
	manifest := schema1{
		FSLayers: []fslayer{{
			BlobSum: digest.String(),
		}},
	}
	b, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	desc := &remote.Descriptor{
		Manifest: b,
		Descriptor: v1.Descriptor{
			MediaType: types.DockerManifestSchema1,
		},
	}
	if err := putManifest(desc.Manifest, desc.MediaType, dstRef, authn.Anonymous); err != nil {
		t.Fatal(err)
	}

	if err := CopySchema1(desc, srcRef, dstRef, authn.Anonymous, authn.Anonymous); err != nil {
		t.Errorf("failed to copy schema 1: %v", err)
	}
}
