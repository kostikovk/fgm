# Decision Response Example

Optional layout example; the [Response Format](../SKILL.md#response-format) defines required evidence and proportional detail.

```markdown
**Recommendation:** <layer / style / structure in one line>

**Why this and not <runner-up>:** <the single deciding factor>

**Source:** <book/topic with a substantiated chapter if available, or official spec/documentation; disclose an unverified attribution>

**Project evidence:** <observed path + symbol + relevant behavior, or identified supplied snippet; state unavailable context>

**How it fits the project:** <what stays, what is reused or extended, and the smallest sufficient change; explain any departure from a precedent or docs>

**Trade-offs** (only for architecture-level calls — use the table in `architecture-selector.md`)

**Testing:** <unit / integration / e2e — which layer gets which>

**Watch out for:** <concurrency, idempotency, error translation, tenant scoping — only the ones that apply>
```
