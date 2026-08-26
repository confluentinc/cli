---
paths:
  - "internal/**/*.go"
---

# Generator alignment and internal-reference hygiene

Much of the CMF/CP command surface is emitted by `cli-terraform-generator`, and the standing goal is
to hand-write as little as possible.

- Flag hand-written Go that only re-derives a value already present in the OpenAPI spec, e.g. a
  hardcoded pending/terminal phase set that duplicates the status enum, or a default hardcoded in the
  command that the spec already defines.
- Flag a one-off helper (map-field extraction, output conversion) that duplicates a sibling instead
  of being generalized into `pkg/flink` (or similar). (Deeper "should this whole thing be generated?"
  judgment is for human reviewers, not a per-diff check.)

Internal-reference hygiene:

- No internal identifiers (ticket keys, RFC links, internal service or system names) should leak into
  user-facing strings: `Short`/`Long`/`Example`, errors, suggestions. They may stay in engineer-facing
  doc comments if useful there.
- Flag PR-introduced comments that carry a stale ticket reference or sit above the wrong function.
