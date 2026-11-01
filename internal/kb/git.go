package kb

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ContractorKeith/mdkb/internal/config"
)

func inGit(root string) bool {
	return exec.Command("git", "-C", root, "rev-parse", "--show-toplevel").Run() == nil
}

func initGit(root string) error {
	if inGit(root) {
		return nil
	}
	output, err := exec.Command("git", "-C", root, "init", "--quiet").CombinedOutput()
	if err != nil {
		return fmt.Errorf("git init: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

// CommitWrite stages and commits only the supplied root-relative paths.
func CommitWrite(root string, paths []string, message string) error {
	cfg, err := config.Writes()
	if err != nil {
		return err
	}
	if !cfg.CommitOnWrite || !inGit(root) || len(paths) == 0 {
		return nil
	}

	// A moved source that was never tracked has no deletion to stage.
	filtered := []string{}
	for _, path := range paths {
		if _, err := os.Lstat(filepath.Join(root, path)); os.IsNotExist(err) {
			if exec.Command("git", "-C", root, "--literal-pathspecs", "ls-files", "--error-unmatch", "--", path).Run() != nil {
				continue
			}
		}
		filtered = append(filtered, path)
	}
	paths = filtered
	if len(paths) == 0 {
		return nil
	}
	args := append([]string{"-C", root, "--literal-pathspecs", "add", "--"}, paths...)
	if output, err := exec.Command("git", args...).CombinedOutput(); err != nil {
		return fmt.Errorf("git add: %w: %s", err, strings.TrimSpace(string(output)))
	}
	diff := append([]string{"-C", root, "--literal-pathspecs", "diff", "--cached", "--quiet", "--"}, paths...)
	if err := exec.Command("git", diff...).Run(); err == nil {
		return nil
	}
	args = []string{"-C", root, "--literal-pathspecs"}
	if output, err := exec.Command("git", "-C", root, "config", "user.email").Output(); err != nil || strings.TrimSpace(string(output)) == "" {
		args = append(args, "-c", "user.name=mdkb", "-c", "user.email=mdkb@local")
	}
	args = append(args, "commit", "-q", "-m", message, "--")
	args = append(args, paths...)
	if output, err := exec.Command("git", args...).CombinedOutput(); err != nil {
		return fmt.Errorf("git commit: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

// CommitWriteStatus reports whether CommitWrite actually advanced HEAD.
func CommitWriteStatus(root string, paths []string, message string) (bool, error) {
	revision := func() string {
		out, _ := exec.Command("git", "-C", root, "rev-parse", "--verify", "HEAD").Output()
		return string(out)
	}
	before := revision()
	if err := CommitWrite(root, paths, message); err != nil {
		return false, err
	}
	return revision() != before, nil
}
