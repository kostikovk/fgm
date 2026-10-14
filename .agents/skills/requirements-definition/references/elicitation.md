# Product request elicitation

Use this procedure when a request mixes problems, solution suggestions and missing decisions. Work from the request, available stakeholders' evidence, existing product documentation and relevant repository behavior. Do not invent interviews, usage data or authority that is unavailable.

## Recover the need

Identify the affected people and circumstances, the current difficulty, and the observable improvement they need. Separate the requested outcome from suggested UI, architecture or tooling. A proposed solution may be a genuine constraint; record whether it is mandated or merely suggested before offering alternatives.

Capture the source and rationale for important rules. Distinguish a reported pain point from an observed fact; a metric or success threshold without supporting authority remains a proposal. Preserve disagreements instead of merging incompatible expectations into a seemingly settled requirement.

Check scope against realistic scenarios: actors, permissions, inputs, transitions, results and interactions with existing behavior. Consider relevant failure, recovery, compatibility and data-integrity cases. Select these from the domain and request; do not apply a universal security or performance checklist to every change.

## Map rules with examples

For each meaningful rule, find a concrete situation and its expected result. Example Mapping's distinction between rules, examples and questions is useful here:

- Rule: only an authorized administrator can issue a team invitation.
- Example: an administrator invites a known account and receives a usable invitation.
- Boundary example: a regular member attempts the same action; no invitation is created.
- Question: can an administrator invite someone who has no account?

Do not decide the last question through an example. If stakeholders have supplied conflicting expectations, show the conflict and its impact. Examples establish observable acceptance; they need not prescribe test frameworks or internal implementation.

Ask focused questions only where missing information materially changes scope, correctness or acceptance. Group related questions when they share a decision, but avoid a broad interview that delays an already clear independent change. State who or what can resolve the question when known; otherwise record the missing decision source.

## Close discovery for the selected scope

Use the [requirement contract](requirement-contract.md) to record the settled behavior, unresolved questions, deferred scenarios and readiness. Resolve enough for the next coherent delivery scope, not every future possibility. If a blocking question cannot be answered, identify the bounded investigation or product decision needed and continue only unaffected ready work.

## Sources and adaptation

[SEBoK — Stakeholder Needs Definition](https://sebokwiki.org/wiki/Stakeholder_Needs_Definition) informs the separation of needs, intended use and stakeholder perspectives. [Cucumber — Introducing Example Mapping](https://cucumber.io/blog/bdd/example-mapping-introduction/) informs rules/examples/questions. The separation of AI proposals from confirmed requirements and the scoped handoff are this repository's adaptation for agent execution; they are not source-mandated document formats.
