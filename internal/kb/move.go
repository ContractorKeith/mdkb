package kb

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type Rewritten struct {
	ID    string `json:"id"`
	Count int    `json:"count"`
}
type MoveResult struct {
	Old         string      `json:"old"`
	New         string      `json:"new"`
	Rewritten   []Rewritten `json:"rewritten"`
	Total       int         `json:"total"`
	Commit      bool        `json:"commit"`
	DryRun      bool        `json:"dry_run,omitempty"`
	CommitError string      `json:"-"`
}
type MovePlan struct {
	Result  MoveResult
	Changes []MoveChange
}
type MoveChange struct{ ID, Before, After string }

// PlanMove validates every affected path before any mutation.
func PlanMove(root, old, new string, sources []string) (MovePlan, error) {
	plan := MovePlan{Result: MoveResult{Old: old, New: new, Rewritten: []Rewritten{}}}
	if ValidateID(old, false) != nil || ValidateID(new, true) != nil {
		return plan, refuse(fmt.Errorf("invalid note id"))
	}
	if err := PreviewWrite(root, old, "overwrite"); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return plan, NoteNotFound(old)
		}
		return plan, err
	}
	if err := PreviewWrite(root, new, "create"); err != nil {
		return plan, err
	}
	_, raw, err := Read(root, old)
	if err != nil {
		return plan, err
	}
	notes, err := Walk(root)
	if err != nil {
		return plan, err
	}
	after := append([]Note(nil), notes...)
	for i := range after {
		if after[i].ID == old {
			after[i].ID = new
		}
	}
	// Include the moved note for self-links even if the caller omitted it.
	seen := map[string]bool{}
	sources = append(append([]string(nil), sources...), old)
	for _, id := range sources {
		if seen[id] {
			continue
		}
		seen[id] = true
		note, content, err := Read(root, id)
		if err != nil {
			return plan, err
		}
		body, count := RewriteLinks(note.Body, old, new, notes, after)
		if count == 0 {
			continue
		}
		if err := PreviewWrite(root, id, "overwrite"); err != nil {
			return plan, err
		}
		replacement := content[:len(content)-len(note.Body)] + body
		if !strings.HasSuffix(replacement, "\n") {
			replacement += "\n"
		}
		plan.Changes = append(plan.Changes, MoveChange{id, content, replacement})
		plan.Result.Rewritten = append(plan.Result.Rewritten, Rewritten{id, count})
		plan.Result.Total += count
	}
	plan.Changes = append([]MoveChange{{old, raw, raw}}, plan.Changes...)
	return plan, nil
}

// ApplyMove preserves the original bytes, then routes changed sources through Write.
func ApplyMove(root string, plan MovePlan) error {
	writeMu.Lock()
	err := renameNote(root, plan.Result.Old, plan.Result.New)
	writeMu.Unlock()
	if err != nil {
		return refuse(err)
	}
	for _, change := range plan.Changes[1:] {
		id := change.ID
		if id == plan.Result.Old {
			id = plan.Result.New
		}
		if _, err := Write(root, id, change.After, "overwrite"); err != nil {
			return err
		}
	}
	return nil
}
