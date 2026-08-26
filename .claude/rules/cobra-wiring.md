---
paths:
  - "internal/**/command*.go"
---

# Cobra command wiring and Cloud/On-Prem gating

- A subcommand must be registered on its parent (new top-level commands in `internal/command.go`);
  an unregistered command file silently does not exist.
- `Args` matches arity: `cobra.ExactArgs(N)` for fixed arity, `cobra.MinimumNArgs(1)` for
  variadic/`delete`. Set `ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs)` where completion
  applies.
- Mode-specific commands gate via the `annotations.RunRequirement` annotation
  (`RequireCloudLogin` / `RequireOnPremLogin` / etc.) with the correct requirement. Meaningfully
  divergent Cloud vs. on-prem behavior belongs in a `command_<sub>_onprem.go` sibling, not an
  `if cfg.IsCloudLogin() { ... } else { ... }` branch inside one handler.
- Multi-resource delete routes confirmation and deletion through `deletion.ValidateAndConfirm` and
  `deletion.Delete`, not a hand-rolled prompt/loop; delete accepts variadic args.
- Handlers access clients lazily via `c.V2Client` / `c.MDSClient` / `c.GetKafkaREST()`. Flag any
  handler that authenticates or constructs a client itself; that belongs in the shared `PreRunner`.
