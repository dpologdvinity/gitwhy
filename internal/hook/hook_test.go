package hook

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dpologdvinity/gitwhy/internal/snapshot"
	"github.com/dpologdvinity/gitwhy/internal/testrepo"
)

func TestPreSkips(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("base", "f.txt", "1\n")
	r.Write("f.txt", "dirty\n")
	cases := map[string][]string{
		t.TempDir(): {"reset", "--hard"},
		r.Top:       {"-C", "x", "reset", "--hard"},
	}
	for dir, args := range cases {
		if id := Pre(dir, args); id != "" {
			t.Errorf("Pre(%s, %v) = %q", dir, args, id)
		}
	}
	for _, args := range [][]string{{"status"}, {"push", "--force"}, {}} {
		if id := Pre(r.Top, args); id != "" {
			t.Errorf("Pre(%v) = %q", args, id)
		}
	}
	t.Setenv("GITWHY_HOOK_OFF", "1")
	if id := Pre(r.Top, []string{"reset", "--hard"}); id != "" {
		t.Errorf("hook off: %q", id)
	}
}

func TestPreThenUndo(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("base", "f.txt", "1\n")
	r.Write("f.txt", "dirty\n")
	id := Pre(r.Top, []string{"reset", "--hard"})
	if id == "" {
		t.Fatal("no snapshot for reset --hard")
	}
	r.Git("reset", "--hard")
	Post(r.Top, id)
	s, err := snapshot.Get(r.Repo, id)
	if err != nil {
		t.Fatal(err)
	}
	if s.After == nil {
		t.Fatal("Post did not finish the snapshot")
	}
	if _, err := snapshot.Restore(r.Repo, s); err != nil {
		t.Fatal(err)
	}
	if got := r.Read("f.txt"); got != "dirty\n" {
		t.Fatalf("f.txt = %q", got)
	}
}

func TestHookScriptKeepsExitCode(t *testing.T) {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		t.Skip("no git")
	}
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("no bash")
	}
	script, err := Script("bash")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	f := filepath.Join(dir, "t.sh")
	if err := os.WriteFile(f, []byte(script+"\ngit rev-parse --verify nope; echo rc=$?\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	c := exec.Command(bash, f)
	c.Dir = dir
	c.Env = append(os.Environ(), "PATH="+filepath.Dir(gitPath)) // gitwhy not on PATH
	out, _ := c.CombinedOutput()
	if !strings.Contains(string(out), "rc=128") {
		t.Fatalf("output:\n%s", out)
	}
	if _, err := Script("fish"); err == nil || !strings.Contains(err.Error(), "only bash and zsh") {
		t.Fatalf("fish err = %v", err)
	}
}

func TestInstallIdempotentAndUninstall(t *testing.T) {
	rc := filepath.Join(t.TempDir(), ".bashrc")
	if err := os.WriteFile(rc, []byte("export A=1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for i, want := range []bool{true, false} {
		changed, err := Install(rc, "bash")
		if err != nil || changed != want {
			t.Fatalf("install %d: changed=%v err=%v", i, changed, err)
		}
	}
	data, _ := os.ReadFile(rc)
	if n := strings.Count(string(data), "# >>> gitwhy hook >>>"); n != 1 || !strings.Contains(string(data), `eval "$(gitwhy hook init bash)"`) {
		t.Fatalf("rc:\n%s", data)
	}
	if changed, err := Uninstall(rc); err != nil || !changed {
		t.Fatalf("uninstall: %v %v", changed, err)
	}
	if data, _ := os.ReadFile(rc); string(data) != "export A=1\n" {
		t.Fatalf("rc after uninstall = %q", data)
	}
	if changed, _ := Uninstall(rc); changed {
		t.Fatal("second uninstall changed the file")
	}
}
