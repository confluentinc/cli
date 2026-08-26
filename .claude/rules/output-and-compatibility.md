---
paths:
  - "internal/**/*.go"
  - "test/fixtures/output/**"
---

# Output structs and backward compatibility

An `*Out` struct and its `human:` / `serialized:` (or `json:`/`yaml:`) tags are a command's
user-facing contract, and the highest-traffic review surface in this repo. Scrutinize their design.

Output and serialization design:

- Serialized field names mirror the API/spec, not an invented CLI name. Matching the spec is what
  lets these commands be generated later; flag a divergent name.
- `omitempty` belongs on optional response fields only. A field the API always returns should not
  carry it; an optional one should, so it drops out cleanly when absent.
- Use `human:"-"` to hide a field from the human table while keeping it in `-o json`/`-o yaml`,
  rather than dropping the field from the struct. Long or unbounded values usually warrant `human:"-"`.
- The field set should be consistent across human/json/yaml, or an intentional divergence should be
  explained. "A different set of fields per format" is a legitimate thing to question.
- Choose new table columns deliberately: start narrow, since widening later is easy but removing a
  column is a breaking change.

Backward compatibility (this binary ships to enterprise customers; breaking changes need a major release):

- Do not rename or remove a command, subcommand, or flag outside a major release.
- Do not change or remove a `serialized:` field name or the `-o json`/`-o yaml` output format, and do
  not add `omitempty` to a field that previously always emitted.
- Renaming a `human:` column header is NOT breaking; renaming a `serialized:` key IS. Keep that
  distinction straight before flagging.
- Printing a status line (e.g. `Updated "x".`) to stdout on a command that also emits
  `-o json`/`-o yaml` breaks `| jq`. Status messages belong on stderr, or are omitted for resources
  that print a payload.
- A changed json/yaml golden under `test/fixtures/output/**` signals a serialized-surface change;
  confirm it is intended and non-breaking (or gated on a major).
- Features marked EA (Early Access) or OP (Open Preview) may break across minor versions; that
  instability should be disclosed in the command's `Short`/`Long`.
