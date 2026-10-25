package hook

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/dpologdvinity/gitwhy/internal/snapshot"
	"github.com/dpologdvinity/gitwhy/internal/testrepo"
)

// TestPreSkips checks that Pre takes no snapshot outside a repository, for a "-C" folder that does
// not exist, for safe commands, for push (a local snapshot cannot undo it), for an empty command
// line, or when GITWHY_HOOK_OFF=1. A needless snapshot costs time on every git command.
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

// TestPreThenUndo checks the whole hook round trip: Pre snapshots before a plain "git reset
// --hard", Post records the state after it, and restoring the snapshot brings the lost edit back.
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
	s, err := snapshot.Get(r.Repo, idOf(id))
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

// TestHookScriptKeepsExitCode checks that the bash function still runs git and returns git's own
// exit code when gitwhy is not on PATH, so the hook can never break git or scripts that check its
// status. It also checks that shells other than bash and zsh are refused.
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

// TestInstallIdempotentAndUninstall checks that installing twice adds the block once, and that
// uninstalling gives back the rc file exactly as it was, with a second uninstall changing nothing.
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

// TestPreSavesIgnoredFilesFromSubfolder checks that "git clean -fX" run from a subfolder saves the
// ignored file it deletes, so undo brings it back. The preview runs in the subfolder but the
// snapshot reads root-relative paths, so a mix-up there would lose the file.
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
	s, err := snapshot.Get(r.Repo, idOf(id))
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

// TestPreSkipsBareRepo checks that Pre does nothing in a bare repository, which has no working
// tree to snapshot.
func TestPreSkipsBareRepo(t *testing.T) {
	bare := t.TempDir()
	if out, err := exec.Command("git", "init", "-q", "--bare", bare).CombinedOutput(); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if id := Pre(bare, []string{"reset", "--hard"}); id != "" {
		t.Fatalf("Pre in bare repo = %q", id)
	}
}

// TestPreGivesUpQuietlyWhenSnapshotFails checks that when the snapshot cannot be taken (here a
// leftover index.lock) Pre returns no ID instead of failing, so the user's git command still runs.
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

// TestHookScriptWorksInZsh checks that the zsh function loads and keeps git's exit code when
// gitwhy is not on PATH, the same promise TestHookScriptKeepsExitCode makes for bash.
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

// TestUninstallDamagedBlockErrors checks that a begin marker without an end marker is an error,
// rather than Uninstall deleting everything after the marker.
func TestUninstallDamagedBlockErrors(t *testing.T) {
	rc := filepath.Join(t.TempDir(), ".bashrc")
	if err := os.WriteFile(rc, []byte("export A=1\n# >>> gitwhy hook >>>\neval x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Uninstall(rc); err == nil || !strings.Contains(err.Error(), "end marker") {
		t.Fatalf("err = %v", err)
	}
}

// TestUninstallEndMarkerWithoutNewline checks that a block at the very end of the file, with no
// newline after the end marker, is still removed cleanly.
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

// TestRCPath checks which startup file the hook goes in: ~/.bashrc for bash, and for zsh
// $ZDOTDIR/.zshrc when ZDOTDIR is set, else ~/.zshrc.
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
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".bash_profile"), []byte("# profile"), 0o644); err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(dir, ".bashrc")
	if runtime.GOOS == "darwin" {
		want = filepath.Join(dir, ".bash_profile")
	}
	if got := RCPath("bash", dir); got != want {
		t.Fatalf("bash with .bash_profile = %q, want %q", got, want)
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

// TestHookScriptOnlySpawnsGitwhyForRiskyCommands checks that the bash function starts gitwhy only
// for "reset --hard", not for --version, status, log, add or push. Starting gitwhy for every git
// command would slow down everyday use.
func TestHookScriptOnlySpawnsGitwhyForRiskyCommands(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the fake gitwhy is a sh script; Git Bash PATH handling differs. The shell logic is covered on Linux")
	}
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

// runShell runs body as a script in the given shell with a fake gitwhy on PATH (see fakeGitwhy),
// starting in dir, and returns the shell's combined output and the fake gitwhy's call log.
func runShell(t *testing.T, shell, dir, body string) (output, calls string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the fake gitwhy is a sh script; Git Bash PATH handling differs")
	}
	bin, err := exec.LookPath(shell)
	if err != nil {
		t.Skipf("no %s", shell)
	}
	path, log := fakeGitwhy(t)
	script, err := Script(shell)
	if err != nil {
		t.Fatal(err)
	}
	f := filepath.Join(t.TempDir(), "t.sh")
	if err := os.WriteFile(f, []byte(script+"\n"+body+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	c := exec.Command(bin, f)
	c.Dir = dir
	c.Env = append(os.Environ(), "PATH="+path, "HOME="+t.TempDir(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1")
	out, _ := c.CombinedOutput()
	data, _ := os.ReadFile(log)
	return string(out), string(data)
}

// TestPreSnapshotsGlobalOptionForms checks that git's global options in front of the command do
// not hide it from the hook: -C (to the root or a subfolder), --no-pager, -c and -P forms all
// take a snapshot. (--git-dir is refused instead; see TestPreSkipsOptionsThatChangeTheRepository.)
func TestPreSnapshotsGlobalOptionForms(t *testing.T) {
	forms := map[string][]string{
		"-C .":       {"-C", ".", "reset", "--hard"},
		"-C sub":     {"-C", "sub", "reset", "--hard"},
		"--no-pager": {"--no-pager", "reset", "--hard"},
		"-c k=v":     {"-c", "core.pager=cat", "reset", "--hard"},
		"-P":         {"-P", "reset", "--hard"},
		"no options": {"reset", "--hard"},
	}
	for name, args := range forms {
		t.Run(name, func(t *testing.T) {
			r := testrepo.New(t)
			r.Commit("base", "f.txt", "1\n", "sub/g.txt", "g\n")
			r.Write("f.txt", "dirty\n")
			if id := Pre(r.Top, args); id == "" {
				t.Fatalf("Pre(%v) took no snapshot", args)
			}
		})
	}
}

// TestPreSkipsOptionsThatChangeTheRepository checks the documented limit: with --git-dir or
// --work-tree the hook cannot tell which repository the command changes, so it takes no snapshot
// rather than snapshotting the wrong one.
func TestPreSkipsOptionsThatChangeTheRepository(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("base", "f.txt", "1\n")
	r.Write("f.txt", "dirty\n")
	for _, args := range [][]string{
		{"--git-dir=.git", "reset", "--hard"},
		{"--git-dir", ".git", "reset", "--hard"},
		{"--work-tree=.", "reset", "--hard"},
	} {
		if id := Pre(r.Top, args); id != "" {
			t.Errorf("Pre(%v) = %q, want no snapshot", args, id)
		}
	}
}

// TestPreResolvesSimpleAliases checks that a git alias for a risky command is treated as that
// command, while shell aliases (starting with "!") and aliases for safe commands are not.
func TestPreResolvesSimpleAliases(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("base", "f.txt", "1\n")
	r.Git("config", "alias.rh", "reset --hard")
	r.Git("config", "alias.st", "status")
	r.Git("config", "alias.sh", "!sh -c 'git reset --hard'")
	r.Write("f.txt", "dirty\n")
	if id := Pre(r.Top, []string{"rh"}); id == "" {
		t.Error("alias rh = reset --hard took no snapshot")
	}
	if id := Pre(r.Top, []string{"st"}); id != "" {
		t.Errorf("alias st = status took snapshot %q", id)
	}
	if id := Pre(r.Top, []string{"sh"}); id != "" {
		t.Errorf("shell alias took snapshot %q", id)
	}
}

// TestHookFunctionSeesThroughGlobalOptionsAndAliases checks the generated shell function: global
// options before the command, and unknown words (possible aliases), must reach `gitwhy __pre`,
// while safe commands behind the same options must not.
func TestHookFunctionSeesThroughGlobalOptionsAndAliases(t *testing.T) {
	for _, shell := range []string{"bash", "zsh"} {
		t.Run(shell, func(t *testing.T) {
			r := testrepo.New(t)
			r.Commit("base", "f.txt", "1\n")
			body := "git --no-pager reset --hard >/dev/null 2>&1\n" +
				"git -C . reset --hard >/dev/null 2>&1\n" +
				"git -c core.pager=cat reset --hard >/dev/null 2>&1\n" +
				"git rh >/dev/null 2>&1\n" + // not a built-in: could be an alias, so ask gitwhy
				"git --no-pager status >/dev/null 2>&1\n" +
				"git -C . log -1 >/dev/null 2>&1\n" +
				"git -c a=b diff >/dev/null 2>&1\n"
			_, calls := runShell(t, shell, r.Top, body)
			if got := len(strings.Fields(calls)); got != 4 {
				t.Fatalf("%s: gitwhy was called %d times, want 4 (the three reset forms and the unknown word):\n%s", shell, got, calls)
			}
		})
	}
}

// TestHookFunctionLoadsWhenGitIsAliased checks that defining the function works for people who
// already alias git (alias git="git --no-pager", or hub): `git() {` is a parse error then.
func TestHookFunctionLoadsWhenGitIsAliased(t *testing.T) {
	cases := map[string]string{
		"bash": "shopt -s expand_aliases\nalias git='git --no-pager'\n",
		"zsh":  "alias git='git --no-pager'\n",
	}
	for shell, pre := range cases {
		t.Run(shell, func(t *testing.T) {
			bin, err := exec.LookPath(shell)
			if err != nil {
				t.Skipf("no %s", shell)
			}
			script, _ := Script(shell)
			f := filepath.Join(t.TempDir(), "t.sh")
			if err := os.WriteFile(f, []byte(pre+script+"\necho loaded\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			out, _ := exec.Command(bin, f).CombinedOutput()
			if !strings.Contains(string(out), "loaded") || strings.Contains(string(out), "rror") {
				t.Fatalf("%s: the function did not load:\n%s", shell, out)
			}
		})
	}
}

// TestHookFunctionKeepsExitCodeUnderSetU checks that `set -u` (nounset) in bash does not turn a
// successful git into "unbound variable" and exit status 127.
func TestHookFunctionKeepsExitCodeUnderSetU(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("base", "f.txt", "1\n")
	out, _ := runShell(t, "bash", r.Top, "set -u\ngit rev-parse --short HEAD >/dev/null\necho rc=$?\n")
	if !strings.Contains(out, "rc=0") || strings.Contains(out, "unbound") {
		t.Fatalf("under set -u:\n%s", out)
	}
}

// idOf returns the snapshot ID inside the token that Pre returns ("<id>@<repository top>").
func idOf(token string) string {
	id, _, _ := strings.Cut(token, "@")
	return id
}

// TestPostFinishesTheSnapshotInTheRepoPreUsed checks that Post records the "after" state in the
// repository Pre snapshotted even when the shell's folder is elsewhere, as with `git -C other
// reset --hard`. Opening the repository at the shell's folder would silently skip it, and undo
// would then also revert later work.
func TestPostFinishesTheSnapshotInTheRepoPreUsed(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("base", "f.txt", "1\n")
	r.Write("f.txt", "dirty\n")
	elsewhere := t.TempDir() // the shell is not inside the repository
	token := Pre(elsewhere, []string{"-C", r.Top, "reset", "--hard"})
	if token == "" {
		t.Fatal("Pre took no snapshot")
	}
	r.Git("reset", "--hard")
	Post(elsewhere, token)
	s, err := snapshot.Get(r.Repo, idOf(token))
	if err != nil {
		t.Fatal(err)
	}
	if s.After == nil {
		t.Fatal("Post did not finish the snapshot: undo would lack its after-state")
	}
}

// TestCasePatternIsValidWithoutSafeCommands checks that the shell "case" pattern stays valid
// shell even when there are no safe commands to list: a bare "|”" is a syntax error.
func TestCasePatternIsValidWithoutSafeCommands(t *testing.T) {
	if got := casePattern(nil); got != "''" {
		t.Fatalf("casePattern(nil) = %q, want ''", got)
	}
	if got := casePattern([]string{"status", "log"}); got != "status|log|''" {
		t.Fatalf("casePattern = %q", got)
	}
}

// TestPreComposesRelativeDashCOptions checks that several -C options combine the way git does:
// each relative one is taken relative to the previous one, so `-C sub -C ..` is the folder we
// started in, not the parent of it.
func TestPreComposesRelativeDashCOptions(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("base", "f.txt", "1\n", "sub/g.txt", "g\n")
	r.Write("f.txt", "dirty\n")
	if id := Pre(r.Top, []string{"-C", "sub", "-C", "..", "reset", "--hard"}); id == "" {
		t.Fatal("Pre did not resolve `-C sub -C ..` to the repository")
	}
}

// TestPreAttachedDashCOption checks that -C<dir> attached syntax is parsed correctly.
func TestPreAttachedDashCOption(t *testing.T) {
	r := testrepo.New(t)
	r.Commit("base", "f.txt", "1\n")
	r.Write("f.txt", "dirty\n")
	parent := filepath.Dir(r.Top)
	base := filepath.Base(r.Top)
	if id := Pre(parent, []string{"-C" + base, "reset", "--hard"}); id == "" {
		t.Fatal("Pre did not resolve -C<dir> attached option")
	}
}
