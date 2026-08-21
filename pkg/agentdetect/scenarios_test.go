package agentdetect

import "errors"

// Synthetic process trees, so every failure mode can be demonstrated locally
// without installing six different agents.
//
// These drive the same Detect() code path as a live run — only the ProcSource
// and the environment lookup are swapped. Nothing about the detection logic is
// special-cased for them.

type scenario struct {
	Name string
	// Desc is why the scenario exists. It is reported on failure, so a broken
	// expectation says what behaviour was supposed to be demonstrated instead of
	// only which name it was filed under.
	Desc string
	Tree []ProcInfo // index 0 is the CLI's parent, walking upward
	Env  map[string]string
}

// fakeSource is a hand-built tree for tests. It, and the sentinel below, live here
// rather than beside the ProcSource interface so no test fixture ships in the
// customer binary.
type fakeSource map[int]ProcInfo

var errNoSuchProcess = errors.New("no such process")

func (f fakeSource) Info(pid int) (ProcInfo, error) {
	p, ok := f[pid]
	if !ok {
		return ProcInfo{}, errNoSuchProcess
	}
	return p, nil
}

var scenarios = []scenario{
	{
		Name: "direct",
		Desc: "Agent shells out to the CLI. Both signals fire and agree.",
		Tree: chain("zsh", "claude"),
		Env:  map[string]string{"CLAUDECODE": "1"},
	},
	{
		Name: "inherited-env",
		Desc: "Human in a tmux session that inherited CLAUDECODE hours ago. Env var says agent;\n" +
			"         ancestry says otherwise. Env-var-only detection counts this as agent traffic.",
		Tree: chain("zsh", "tmux", "ghostty"),
		Env:  map[string]string{"CLAUDECODE": "1"},
	},
	{
		Name: "ci-image",
		Desc: "CI runner whose image bakes in an agent variable. Same false positive, at scale.",
		Tree: chain("bash", "runner", "dockerd"),
		Env:  map[string]string{"CLAUDECODE": "1", "CI": "true", "GITHUB_ACTIONS": "true"},
	},
	{
		Name: "node-hosted",
		Desc: "Agent distributed as a Node package. Basename is 'node'; only argv identifies it.\n" +
			"         This is the recall gap that name-only matching leaves open.",
		Tree: []ProcInfo{
			{Pid: 100, Ppid: 101, Name: "zsh"},
			{Pid: 101, Ppid: 1, Name: "node", Cmdline: []string{"node", "/usr/local/lib/node_modules/@anthropic-ai/claude-code/cli.js", "--print"}},
		},
	},
	{
		Name: "node-hosted-blind",
		Desc: "Same tree, but argv is unavailable — a cross-user/elevated ancestor, or a platform read\n" +
			"         denied for some other reason. Detection misses entirely. Not Windows-specific:\n" +
			"         argv reads are permission-gated on every platform, and fail the same way.",
		Tree: chain("zsh", "node"),
	},
	{
		Name: "interpreter-gap",
		Desc: "A Node-hosted agent we have no cmdline fingerprint for. Nothing matches, but the\n" +
			"         unattributed interpreter is still counted — that count is how we size the\n" +
			"         recall gap by population, so a table update can be aimed at it.",
		Tree: []ProcInfo{
			{Pid: 100, Ppid: 101, Name: "zsh"},
			{Pid: 101, Ppid: 1, Name: "node", Cmdline: []string{"node", "/opt/somenewagent/dist/main.js"}},
		},
	},
	{
		Name: "pid-reuse",
		Desc: "The parent pid was recycled: it reports a start time AFTER its child's, so it cannot\n" +
			"         really be the parent. Walking on would splice a fabricated ancestor into the chain.",
		Tree: []ProcInfo{
			{Pid: 100, Ppid: 101, Name: "zsh", StartTime: 5000},
			{Pid: 101, Ppid: 1, Name: "claude", StartTime: 9000},
		},
	},
	{
		Name: "unknown-vendor",
		Desc: "An agent we have a process fingerprint for but no env fingerprint (or the vendor\n" +
			"         renamed their variable). Env-var-only detection MISSES this call.",
		Tree: chain("bash", "aider"),
	},
	{
		Name: "wrapper-chain",
		Desc: "Agent → make → npm → CLI. Checking only the immediate parent would miss this.",
		Tree: chain("sh", "make", "npm", "bash", "claude"),
	},
	{
		Name: "timeout-wrapped",
		Desc: "Agent → timeout → CLI. A command being force-timed-out (e.g. one that would\n" +
			"         otherwise run forever); the wrapper's presence is itself the finding.",
		Tree: chain("timeout", "bash", "claude"),
		Env:  map[string]string{"CLAUDECODE": "1"},
	},
	{
		Name: "xargs-fanout",
		Desc: "Agent → xargs -P20 → CLI. A parallel fan-out that can trip server-side rate\n" +
			"         limiting; worth counting because it predicts throttling.",
		Tree: chain("xargs", "bash", "claude"),
	},
	{
		Name: "ide-terminal",
		Desc: "Human typing in an editor's built-in terminal. Agent-capable environment, not an\n" +
			"         agent call. Reported as ide_host, deliberately not as an agent.",
		Tree: chain("zsh", "cursor"),
	},
	// Four in-editor surfaces captured from real process trees: three IDE chat
	// panels and one integrated terminal. Phase 1 reports the same thing for all
	// of them — an editor is in the ancestry (ide_host), plus whatever env var is
	// set. The agent-vs-human distinction is deliberately not attempted; see
	// ide_surfaces_test.go.
	{
		Name: "vscode-claude-chat",
		Desc: "Claude Code's VS Code extension. The extension launches the real CLI as a child of the\n" +
			"         extension host, so it runs as the Electron helper binary — basename\n" +
			"         'code helper (plugin)', never 'node'. That basename is kindIDEHost and not\n" +
			"         argv-eligible, so ancestry cannot name the agent; the CLAUDECODE env var carries it.",
		Tree: []ProcInfo{
			{Pid: 100, Ppid: 101, Name: "zsh"},
			{Pid: 101, Ppid: 102, Name: "code helper (plugin)", Cmdline: []string{
				"/Applications/Visual Studio Code.app/Contents/Frameworks/Code Helper (Plugin).app/Contents/MacOS/Code Helper (Plugin)",
				"/Users/x/.vscode/extensions/anthropic.claude-code-2.1.220-darwin-arm64/resources/claude-code/cli.js",
			}},
			{Pid: 102, Ppid: 103, Name: "code helper (plugin)", Cmdline: []string{
				"/Applications/Visual Studio Code.app/Contents/Frameworks/Code Helper (Plugin).app/Contents/MacOS/Code Helper (Plugin)",
				"--type=utility", "--utility-sub-type=node.mojom.NodeService",
			}},
			{Pid: 103, Ppid: 1, Name: "code"},
		},
		Env: map[string]string{"CLAUDECODE": "1"},
	},
	{
		Name: "vscode-copilot-chat",
		Desc: "Copilot chat in VS Code. The shell's direct parent IS the extension host, whose argv is\n" +
			"         Chromium boilerplate — there is NO agent process, so no fingerprint table can name\n" +
			"         the vendor from ancestry. The editor is reported as ide_host and the COPILOT_CLI env\n" +
			"         var is the only vendor signal: the shape where env-var detection beats the process tree.",
		Tree: []ProcInfo{
			{Pid: 100, Ppid: 101, Name: "bash"},
			{Pid: 101, Ppid: 102, Name: "code helper (plugin)", Cmdline: []string{
				"/Applications/Visual Studio Code.app/Contents/Frameworks/Code Helper (Plugin).app/Contents/MacOS/Code Helper (Plugin)",
				"--type=utility", "--utility-sub-type=node.mojom.NodeService", "--lang=en-US",
			}},
			{Pid: 102, Ppid: 1, Name: "code"},
		},
		Env: map[string]string{"COPILOT_CLI": "1"},
	},
	{
		Name: "cursor-agent-chat",
		Desc: "Cursor's agent. Same in-extension-host shape as Copilot, plus cursorsandbox interposed\n" +
			"         between the shell and the helper. Reported as ide_host (cursor) with the sandbox as a\n" +
			"         wrapper and CURSOR_AGENT as the vendor signal; the rewritten helper argv names nothing.",
		Tree: []ProcInfo{
			{Pid: 100, Ppid: 101, Name: "zsh"},
			{Pid: 101, Ppid: 102, Name: "zsh"},
			{Pid: 102, Ppid: 103, Name: "cursorsandbox"},
			{Pid: 103, Ppid: 104, Name: "cursor helper (plugin)", Cmdline: []string{
				"Cursor Helper (Plugin): extension-host agentdetect [1-2]",
			}},
			{Pid: 104, Ppid: 1, Name: "cursor"},
		},
		Env: map[string]string{"CURSOR_AGENT": "1"},
	},
	{
		Name: "ide-integrated-terminal",
		Desc: "A human typing in VS Code's integrated terminal. In Phase 1 this reports the same ide_host\n" +
			"         as the agent surfaces above and no agent — the agent-vs-human split is not attempted.\n" +
			"         With no env var set, this is correctly silent on the agent question.",
		Tree: []ProcInfo{
			{Pid: 100, Ppid: 101, Name: "zsh"},
			{Pid: 101, Ppid: 102, Name: "code helper", Cmdline: []string{
				"/Applications/Visual Studio Code.app/Contents/Frameworks/Code Helper.app/Contents/MacOS/Code Helper",
				"--type=utility", "--utility-sub-type=node.mojom.NodeService",
			}},
			{Pid: 102, Ppid: 1, Name: "code"},
		},
	},
	{
		Name: "ide-terminal-stale-env",
		Desc: "The same integrated terminal, with a stale agent variable inherited into it. Reports\n" +
			"         ide_host plus the env vendor, with no agent ancestor — the ambiguous case where only\n" +
			"         the env var fired and an editor is in the chain, for downstream analytics to weigh\n" +
			"         rather than something this package resolves.",
		Tree: []ProcInfo{
			{Pid: 100, Ppid: 101, Name: "zsh"},
			{Pid: 101, Ppid: 102, Name: "code helper"},
			{Pid: 102, Ppid: 1, Name: "code"},
		},
		Env: map[string]string{"CLAUDECODE": "1"},
	},
	{
		Name: "nested-agents",
		Desc: "One agent delegating to another. Shallowest wins; the disagreement is preserved\n" +
			"         rather than resolved client-side.",
		Tree: chain("bash", "codex", "bash", "claude"),
		Env:  map[string]string{"CLAUDECODE": "1"},
	},
	{
		Name: "ssh",
		Desc: "Agent on a laptop, CLI on a remote host. Ancestry severed at sshd, env stripped.\n" +
			"         Both signals blind — a shared ceiling, not a tiebreaker.",
		Tree: chain("bash", "sshd"),
	},
	{
		Name: "container",
		Desc: "CLI inside a container the agent exec'd into. PID namespace hides the agent, but a\n" +
			"         forwarded env var survives. The one case where env vars are strictly better.",
		Tree: chain("sh"),
		Env:  map[string]string{"CLAUDECODE": "1"},
	},
	{
		Name: "human",
		Desc: "Ordinary human invocation. Both signals correctly silent.",
		Tree: chain("zsh", "ghostty"),
	},
}

// chain builds a linear ancestry from names, parent-first.
func chain(names ...string) []ProcInfo {
	out := make([]ProcInfo, len(names))
	for i, n := range names {
		ppid := 100 + i + 1
		if i == len(names)-1 {
			ppid = 1
		}
		out[i] = ProcInfo{Pid: 100 + i, Ppid: ppid, Name: n}
	}
	return out
}

// sourceFor indexes a parent-first chain by pid — the shape a ProcSource has to
// be — and returns the pid the walk starts from.
func sourceFor(procs []ProcInfo) (fakeSource, int) {
	src := fakeSource{}
	for _, p := range procs {
		src[p.Pid] = p
	}
	return src, procs[0].Pid
}

func (s scenario) source() (ProcSource, int) {
	return sourceFor(s.Tree)
}

func (s scenario) getenv() func(string) (string, bool) {
	return func(k string) (string, bool) {
		v, ok := s.Env[k]
		return v, ok
	}
}

func findScenario(name string) (scenario, bool) {
	for _, s := range scenarios {
		if s.Name == name {
			return s, true
		}
	}
	return scenario{}, false
}
