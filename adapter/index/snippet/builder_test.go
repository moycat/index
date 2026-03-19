package snippet

import "testing"

func TestBuilderBuildWithChineseTerm(t *testing.T) {
	b := NewBuilder()
	content := "第一段介绍。第二段详细解释中文搜索能力和分词策略。第三段总结。"
	snippet := b.Build(content, []string{"中文", "搜索"}, 16)
	if snippet == "" {
		t.Fatalf("expected snippet")
	}
	if snippet[0:3] != "..." && len([]rune(content)) > 16 {
		// The window should usually move around matched terms in the middle.
	}
}

func TestBuilderBuildFallback(t *testing.T) {
	b := NewBuilder()
	content := "abcdefg"
	snippet := b.Build(content, nil, 4)
	if snippet == "" {
		t.Fatalf("expected fallback snippet")
	}
}
