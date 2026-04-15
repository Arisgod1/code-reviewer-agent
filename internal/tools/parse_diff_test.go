package tools

import "testing"

func TestParseDiffContent(t *testing.T) {
	diff := `diff --git a/a.go b/a.go
+++ b/a.go
@@ -1,1 +1,2 @@
+fmt.Println("hello")
+sql := "select * from t where id=" + id
`
	res := ParseDiffContent(diff)

	if len(res.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(res.Files))
	}
	if res.Files[0] != "a.go" {
		t.Fatalf("expected file a.go, got %s", res.Files[0])
	}
	if len(res.Lines) != 2 {
		t.Fatalf("expected 2 changed lines, got %d", len(res.Lines))
	}
}
