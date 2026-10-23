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

func TestPreSavesIgnoredFilesFromSubfolder(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("base", ".gitignore", "*.log\n", "sub/a.txt", "a\n")
	r.Write("sub/build.log", "keep me\n")
	sub := filepath.Join(r.Top, "sub")
	id := Pre(sub, []string{"clean", "-fX", "build.log"})
	if id == "" {
		t.Fatal("no snapshot")
	}
	r.Git("clean", "-fXq")
	Post(sub, id)
	s, err := snapshot.Get(r.Repo, id)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := snapshot.Restore(r.Repo, s); err != nil {
		t.Fatal(err)
	}
	if got := r.Read("sub/build.log"); got != "keep me\n" {
		t.Fatalf("build.log = %q", got)
	}
}

func TestPreSkipsBareRepo(t *testing.T) {
	bare := t.TempDir()
	if out, err := exec.Command("git", "init", "-q", "--bare", bare).CombinedOutput(); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if id := Pre(bare, []string{"reset", "--hard"}); id != "" {
		t.Fatalf("Pre in bare repo = %q", id)
	}
}

func TestPreGivesUpQuietlyWhenSnapshotFails(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("base", "f.txt", "1\n")
	r.Write("f.txt", "dirty\n")
	if err := os.WriteFile(filepath.Join(r.GitDir, "index.lock"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if id := Pre(r.Top, []string{"reset", "--hard"}); id != "" {
		t.Fatalf("Pre with index.lock = %q", id)
	}
}

func TestHookScriptWorksInZsh(t *testing.T) {
	zsh, err := exec.LookPath("zsh")
	if err != nil {
		t.Skip("no zsh")
	}
	gitPath, err := exec.LookPath("git")
	if err != nil {
		t.Skip("no git")
	}
	script, _ := Script("zsh")
	f := filepath.Join(t.TempDir(), "t.zsh")
	if err := os.WriteFile(f, []byte(script+"\ngit rev-parse --verify nope; echo rc=$?\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	c := exec.Command(zsh, f)
	c.Env = append(os.Environ(), "PATH="+filepath.Dir(gitPath))
	out, _ := c.CombinedOutput()
	if !strings.Contains(string(out), "rc=128") {
		t.Fatalf("output:\n%s", out)

	}
}
func TestUninstallDamagedBlockErrors(t *testing.T) {
	rc := filepath.Join(t.TempDir(), ".bashrc")
	if err := os.WriteFile(rc, []byte("export A=1\n# >>> gitwhy hook >>>\neval x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Uninstall(rc); err == nil || !strings.Contains(err.Error(), "end marker") {
		t.Fatalf("err = %v", err)
	}
}

func TestUninstallEndMarkerWithoutNewline(t *testing.T) {
	rc := filepath.Join(t.TempDir(), ".bashrc")
	if err := os.WriteFile(rc, []byte("export A=1\n# >>> gitwhy hook >>>\neval x\n# <<< gitwhy hook <<<"), 0o644); err != nil {
		t.Fatal(err)
	}
	if changed, err := Uninstall(rc); err != nil || !changed {
		t.Fatalf("changed=%v err=%v", changed, err)
	}
	if data, _ := os.ReadFile(rc); string(data) != "export A=1\n" {
		t.Fatalf("rc = %q", data)
	}
}

func TestRCPath(t *testing.T) {
	t.Setenv("ZDOTDIR", "")
	if got := RCPath("zsh", "/h"); got != filepath.Join("/h", ".zshrc") {
		t.Fatalf("zsh = %q", got)
	}
	t.Setenv("ZDOTDIR", "/z")
	if got := RCPath("zsh", "/h"); got != filepath.Join("/z", ".zshrc") {
		t.Fatalf("zsh with ZDOTDIR = %q", got)
	}
	if got := RCPath("bash", "/h"); got != filepath.Join("/h", ".bashrc") {
		t.Fatalf("bash = %q", got)
	}
}

// fakeGitwhy puts a gitwhy on PATH that logs each call, so tests can see when the hook spawns it.
func fakeGitwhy(t *testing.T) (pathEnv, logFile string) {
	t.Helper()
	bin := t.TempDir()
	logFile = filepath.Join(t.TempDir(), "calls")
	sh := "#!/bin/sh\necho \"$1\" >> " + logFile + "\n"
	if err := os.WriteFile(filepath.Join(bin, "gitwhy"), []byte(sh), 0o755); err != nil {
		t.Fatal(err)
	}
	gitPath, err := exec.LookPath("git")
	if err != nil {
		t.Skip("no git")
	}
	return bin + string(os.PathListSeparator) + filepath.Dir(gitPath), logFile
}

func TestHookScriptOnlySpawnsGitwhyForRiskyCommands(t *testing.T) {
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("no bash")
	}
	script, _ := Script("bash")
	path, calls := fakeGitwhy(t)
	f := filepath.Join(t.TempDir(), "t.sh")
	body := script + "\ngit --version >/dev/null; git status >/dev/null 2>&1; git log -1 >/dev/null 2>&1; git add -A 2>/dev/null; git push 2>/dev/null; git reset --hard >/dev/null 2>&1\n"
	if err := os.WriteFile(f, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	c := exec.Command(bash, f)
	c.Dir = t.TempDir()
	c.Env = append(os.Environ(), "PATH="+path)
	c.CombinedOutput()
	data, _ := os.ReadFile(calls)
	if got := strings.Fields(string(data)); strings.Join(got, ",") != "__pre" {
		t.Fatalf("gitwhy was called with %v, want only __pre for reset", got)
	}
}
