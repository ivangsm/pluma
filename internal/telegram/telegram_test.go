package telegram

import "testing"

func TestEscapeHTML(t *testing.T) {
	t.Run("ampersand", func(t *testing.T) {
		if got := escapeHTML("&"); got != "&amp;" {
			t.Fatalf("expected '&amp;', got %q", got)
		}
	})

	t.Run("less than", func(t *testing.T) {
		if got := escapeHTML("<"); got != "&lt;" {
			t.Fatalf("expected '&lt;', got %q", got)
		}
	})

	t.Run("greater than", func(t *testing.T) {
		if got := escapeHTML(">"); got != "&gt;" {
			t.Fatalf("expected '&gt;', got %q", got)
		}
	})

	t.Run("double quote", func(t *testing.T) {
		if got := escapeHTML(`"`); got != "&quot;" {
			t.Fatalf("expected '&quot;', got %q", got)
		}
	})

	t.Run("single quote", func(t *testing.T) {
		if got := escapeHTML("'"); got != "&#39;" {
			t.Fatalf("expected '&#39;', got %q", got)
		}
	})

	t.Run("all special characters", func(t *testing.T) {
		input := `<b>"hello" & 'world'</b>`
		expected := "&lt;b&gt;&quot;hello&quot; &amp; &#39;world&#39;&lt;/b&gt;"
		if got := escapeHTML(input); got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	})

	t.Run("normal text unchanged", func(t *testing.T) {
		input := "Hello World 123"
		if got := escapeHTML(input); got != input {
			t.Fatalf("expected %q, got %q", input, got)
		}
	})
}
