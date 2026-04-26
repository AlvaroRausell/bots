package document

import (
	"reflect"
	"testing"
)

func TestUpsertSectionReplacesOnlyTargetSectionOutsideCodeBlocks(t *testing.T) {
	content := "# Title\n\n```\n## Current Checkpoint\nnot a section\n```\n\n## Current Checkpoint\n\nold\n\n## Open Questions\n\nnone\n"

	got := UpsertSection(content, "## Current Checkpoint", "new")
	want := "# Title\n\n```\n## Current Checkpoint\nnot a section\n```\n\n## Current Checkpoint\n\nnew\n\n## Open Questions\n\nnone\n"

	if got != want {
		t.Fatalf("unexpected document:\n--- got ---\n%q\n--- want ---\n%q", got, want)
	}
}

func TestUpsertSectionAppendsMissingSection(t *testing.T) {
	got := UpsertSection("# Title\n", "## Current Checkpoint", "new")
	want := "# Title\n\n## Current Checkpoint\n\nnew\n"

	if got != want {
		t.Fatalf("unexpected document:\n--- got ---\n%q\n--- want ---\n%q", got, want)
	}
}

func TestListSectionsIgnoresHeadingsInsideCodeBlocks(t *testing.T) {
	content := "# Title\n\n## One\n\n```\n## Ignored\n```\n\n## Two\n"

	got := ListSections(content, 2)
	want := []string{"One", "Two"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestSectionBodyStopsAtNextSameOrHigherHeading(t *testing.T) {
	content := "# Title\n\n## Status\n\nPENDING\n\n### Detail\n\nkept\n\n## Other\n\nskipped\n"

	body, ok := SectionBody(content, "## Status")
	if !ok {
		t.Fatal("expected section body to be found")
	}

	want := "PENDING\n\n### Detail\n\nkept"
	if body != want {
		t.Fatalf("expected %q, got %q", want, body)
	}
}
