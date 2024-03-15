package route

import "testing"

func TestRewritePrefix(t *testing.T) {
	r := &Route{
		Name:     "Home",
		Path:     "/",
		PathType: "prefix",
	}

	paths := [][]string{
		{"/", "/"},
		{"/abc", "/abc"},
		{"/a/b/c", "/a/b/c"},
	}

	for _, p := range paths {
		request := p[0]
		expected := p[1]
		if expected != r.Rewrite(request) {
			t.Fatalf("expected %s, but got %s", expected, r.Rewrite(request))
		}
	}
}
