---
paths:
  - "internal/**/*.go"
---

# Backend-first validation and error-handling semantics

Prefer backend validation:

- Let the backend validate spec-recorded constraints (allowed enum values, required-ness, formats)
  and surface its error. Flag new hand-rolled client-side validation that duplicates a spec constraint.
- Use Cobra built-ins for flag relationships: `cmd.MarkFlagsMutuallyExclusive(...)` for `oneOf`
  groups, `cmd.MarkFlagsRequiredTogether(...)` for co-required flags, not a bespoke
  `if flagA != "" && flagB != "" { ... }` ladder.
- Case normalization (`strings.ToUpper` on an enum before sending) should be a conscious choice; for
  `x-extensible-enum` fields a client-side allow-list breaks forward compatibility.
- Do not branch on backend error wording. String-matching a server message is brittle the moment the
  backend rewords, and usually means the validation should have stayed server-side.

Error-handling semantics (casing and `NewErrorWithSuggestions` are owned by convention and
`lint-cli`; spend review attention here instead):

- An existence or precheck helper must not collapse every failure (401, 500, timeout, connection
  refused) into "not found". Only a true 404 maps to "not found"; other errors preserve their cause.
- Status-code gating needs the right boundary: `< 400` treats a 3xx redirect as success and
  unmarshals into a zero-value struct (printing a blank "success"). Prefer an explicit 2xx check,
  noting `201` on create.
- Errors must not be silently swallowed. The underlying cause should reach the user rather than a
  generic substitute message.
- Auth and configuration errors should name the missing or invalid setting.
