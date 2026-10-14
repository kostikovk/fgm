---
name: developer-readable-output
description: Make developer-facing explanations, progress updates, reports, reviews, and plans easy to understand across frontend and backend specialties without losing technical evidence. Use when presenting technical work or when the user asks for clearer, more readable output. Formats communication; does not replace engineering workflows or change required machine-readable output.
---

# Developer Readable Output

Help the developer find the answer, understand its basis, and see whether anything remains to be done. Apply this as a presentation layer alongside the skill doing the work.

## Shared boundaries — always apply

- Start with the useful answer, observed result, recommendation, or blocker. Lead with an action when the reader actually needs to perform it. Avoid introductory filler and retelling the conversation.
- Write for a capable developer who may not specialize in the affected stack. Prefer familiar words and explain necessary technical terms at first use through what the code does or what can go wrong. Keep exact identifiers, but do not use acronyms, pattern names, or claims such as “best practice” as explanations by themselves.
- For a substantive change or decision, make clear what changed, why this choice addresses the observed problem, and which concrete facts support it. Connect code, logs, or test results to a consequence the reader can understand. Distinguish evidence from assumptions; simplify the wording without weakening the technical claim. A small answer does not need a formal report template.
- Preserve the task's required information, evidence, uncertainty, and material limitations. A compact report must not turn an unrun check into a passing check, a hypothesis into a cause, or a partial result into completion. The owning engineering skill still determines its required fields and workflow.
- Follow the user's language, requested detail, and output format. Keep code identifiers and commands exact. A required JSON/schema-only response, patch, exact quotation, or other strict artifact takes precedence over prose formatting; add no conversational wrapper.
- Use short paragraphs with one main point. Use bullets for parallel facts, numbers for ordered actions, and tables for comparisons when they improve scanning. Add headings only when they help navigate a longer response. Avoid nested lists, decorative emphasis, and rigid list or word limits that hide necessary information.
- Keep locations, commands, and evidence near the claim they support. Use readable file links when the surface supports them; otherwise use precise paths and line numbers. Put runnable commands or code in fenced blocks and distinguish examples from commands actually executed.
- Mention a next step only when work remains, with a clear owner and concrete purpose. Continue authorized agent work rather than handing it back as advice. A completed task can end with its result and validation. Do not invent estimates, require a question at the end, or assume a diagnosis or reading preference from this skill's inspiration.

## Navigation — choose the response

For a short answer, the shared boundaries are enough. For a substantial response, select the relevant section below. A fragment locates a heading; use a bounded read through the next heading of the same or higher level, or load only the relevant reference file if section reads are unavailable. Follow the owning task's prerequisites; do not load every route.

| Response needed | Read |
|---|---|
| Explain code, a concept, or an open design choice | [Explanations and decisions](references/developer-responses.md#explanations-and-decisions) |
| Explain unfamiliar frontend/backend behavior or translate dense terminology | [Cross-stack explanations](references/cross-stack-explanations.md) |
| Report completed implementation or validation | [Implementation reports](references/developer-responses.md#implementation-reports) |
| Share progress, a debugging result, or a blocker | [Progress and debugging](references/developer-responses.md#progress-and-debugging) |
| Present review findings | [Code reviews](references/developer-responses.md#code-reviews) |
| Give steps the reader needs to execute | [Action plans](references/developer-responses.md#action-plans) |
| Check the source of this preference and adaptation choices | [Inspiration and scope](references/inspiration.md) |

## Before sending

Check that the opening answers the request, the important evidence and limitations remain visible, and the structure fits the response rather than a fixed template. Could a developer outside this specialty explain what changed and why from this response? Replace unexplained jargon with its concrete meaning. Remove repeated conclusions and optional tangents. If the task is incomplete, state what remains; if it is complete, do not manufacture another task.
