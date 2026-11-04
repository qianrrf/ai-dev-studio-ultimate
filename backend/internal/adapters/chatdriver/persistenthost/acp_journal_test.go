package persistenthost

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/aoagents/agent-orchestrator/backend/internal/testutil"
)

func TestACPPromptJournalReplaysResetsAndEnforcesQuota(t *testing.T) {
	path := filepath.Join(t.TempDir(), "prompt.journal")
	journal, err := openACPPromptJournal(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = journal.close(context.Background()) })
	testutil.AssertPrivatePath(t, path, 0o600)
	var spans []acpJournalSpan
	for _, frame := range [][]byte{[]byte("one\n"), []byte("two\n")} {
		span, err := journal.append(context.Background(), frame)
		if err != nil {
			t.Fatal(err)
		}
		spans = append(spans, span)
	}
	var replay bytes.Buffer
	for _, span := range spans {
		if err := journal.replayTo(context.Background(), &replay, span); err != nil {
			t.Fatal(err)
		}
	}
	if replay.String() != "one\ntwo\n" {
		t.Fatalf("replay = %q", replay.String())
	}
	if err := journal.reset(context.Background()); err != nil {
		t.Fatal(err)
	}
	replay.Reset()
	if err := journal.replayTo(context.Background(), &replay, acpJournalSpan{}); err != nil || replay.Len() != 0 {
		t.Fatalf("reset replay = %q, err=%v", replay.String(), err)
	}
	journal.size = maxACPJournalBytes
	if _, err := journal.append(context.Background(), []byte("overflow\n")); !errors.Is(err, errACPJournalFull) {
		t.Fatalf("quota error = %v", err)
	}
}
