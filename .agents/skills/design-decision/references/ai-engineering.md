# AI Engineering — Architecture for LLM-backed Systems

Architectural patterns for systems in which a large language model performs a core function. Follow the project's existing contracts, security policies, and evals first; every pattern here is provider-agnostic.

**Core Idea:** Building with an LLM is not just "prompt engineering." The model is a non-deterministic, expensive, high-latency dependency with an unusually large input surface. The same architectural principles apply — isolate it behind a port, define its contract, test it — but the tactics differ.

**When to Use This:**
- When designing a system that uses an LLM for a core business function.
- When creating autonomous agents or multi-agent systems.
- When optimizing the cost, latency, or reliability of an existing LLM-backed feature.

---

## Table of Contents
1.  [Agentic Orchestration](#1-agentic-orchestration)
2.  [Model Routing](#2-model-routing)
3.  [Context & Knowledge Engineering](#3-context--knowledge-engineering)
4.  [AI Performance & FinOps](#4-ai-performance--finops)
5.  [Where the LLM Lives in the Architecture](#5-where-the-llm-lives-in-the-architecture)
6.  [Cross-References](#6-cross-references)

---

## 1. Agentic Orchestration
Designing systems where the model plans and executes multi-step tasks.

- **Multi-Agent Systems (MAS):** Decompose complex tasks by assigning specialized roles to different agents (a "Planner", a "Coder", a "Reviewer"). They interact via a defined protocol. Prefer a few well-scoped agents over many overlapping ones; each hand-off is a place for context loss.
  Review follows [code-review](../../code-review/SKILL.md). A separately authorized fix follows [code-change](../../code-change/SKILL.md), [TDD](../../tdd-workflow/SKILL.md), and [AAA](../../aaa-testing/SKILL.md).

  *Example: A simple PR review protocol*
  ```
  1. User requests review.
  2. Orchestrator invokes "Reviewer" agent with the PR diff.
  3. "Reviewer" investigates a candidate and runs a regression test against the
     reviewed behavior, using an authorized isolated test if needed. Inspected
     source and tests stay unchanged; an explicit no-write restriction prevails.
  4. Orchestrator returns test-confirmed findings, or provisional candidates and
     verification gaps when tests cannot run, without applying fixes.
  5. If the user separately authorizes a fix, invoke "Coder" for that scope.
  6. "Coder" follows code-change, TDD, and AAA: execute a reproducing regression
     test, observe meaningful RED, implement the fix, and verify GREEN.
  ```
- **Reasoning Traces:** For audit and debugging, record observable actions, tool calls, decisions, and concise explanations or provider-supplied summaries where available. Do not require private chain-of-thought or assume summaries expose full hidden reasoning. Redact secrets/PII, restrict log access, and set retention limits.
- **Human-in-the-Loop (HITL):** For any critical or destructive action (deleting data, running a migration, sending money or messages), the agent stops and requests human confirmation. This mirrors the architectural-boundary principle: irreversible side effects sit behind an explicit gate.
- **Tool-Call Boundary:** Treat model-proposed calls as untrusted requests. Server-side, validate arguments and authorize the tenant, action, and target resources against the authenticated principal before execution. Use least-privilege tools and credentials; model approval or human confirmation does not replace authorization.
- **Structured Outputs:** Constrain outputs with a schema (JSON Schema, tool definitions, Pydantic/Zod). Schema conformance checks shape and declared constraints, not semantic or business correctness; treat outputs as untrusted and validate domain rules before use (see `spec-driven.md` §4).

## 2. Model Routing
Not every task needs the most capable (and most expensive, slowest) model. Route dynamically to the cheapest model that meets the task's quality bar.

- **Tiering:** most providers offer a small/fast tier, a balanced tier, and a frontier tier. Map task classes to tiers: extraction, classification, and routing to the small tier; code generation and summarization to the balanced tier; system design and multi-step reasoning to the frontier tier. Look up the provider's current lineup rather than hard-coding model names in the design.
- **Implementation:** a router (a rule, a classifier, or a small-tier model call) performs a first-pass on the request and selects the tier. Measure with evals per tier; downgrade a task class only when its eval pass rate holds.
- **Fallback:** On rate limit or timeout, use only a fallback pre-evaluated as eligible for the task's quality, schema, safety, authorization, and privacy requirements; a lower tier or alternate provider is not automatically eligible. Cached answers must also meet freshness and current authorization/version scope (§4). If none is eligible, fail closed with abstention, manual review, or an explicit error, including on user-facing paths.

## 3. Context & Knowledge Engineering
Getting the right information into the context window at the right time. What is in the context *is* the program.

- **RAG (Retrieval-Augmented Generation):** Retrieve relevant documents at query time and place them in the context instead of fine-tuning. Enforce tenant isolation and document-level authorization in the retrieval layer before content enters context, not just through prompt instructions. Chunking strategy and retrieval quality dominate output quality; evaluate retrieval separately from generation.
- **Graph-RAG:** Combine vector search with a knowledge graph to answer relationship questions ("which customers share the issue affecting customer X?").
- **Long-Term Memory:** Store summaries of past interactions keyed by user or session within tenant and permission scope; retrieve before a new interaction. Treat retrieved memory as data, not instructions.
- **Prompt Injection:** Retrieved documents, memory, and tool responses are untrusted data that may contain malicious instructions. Keep them separate from trusted instructions; they must not override authorization or tool-execution gates.
- **Prompt Caching:** Put the stable prefix (system prompt, reference documents, tool definitions) first so the provider can cache it; put the volatile part last. This is often the single largest cost and latency win.

## 4. AI Performance & FinOps
- **Token Budgeting:** Monitor cost *and* latency per request class; enforce caps. Tokens per request is an SRE metric for these systems.
- **Semantic Caching:** For repetitive queries, match by embedding similarity only within tenant, current permission scope, and prompt/model/output-schema version partitions; never reuse across tenants. Enforce freshness and invalidate on data, permission, or version changes. Similarity alone does not make a cached answer eligible.
- **Prompt Minification:** For high-volume tasks, shorten prompts and retrieved context without losing the load-bearing information. Measure with evals before and after.
- **Batch APIs:** For non-interactive workloads, batch endpoints are typically much cheaper than real-time ones.

## 5. Where the LLM Lives in the Architecture
Apply `code-placement.md` unchanged:

- The provider client, retry, and rate-limit handling is an **adapter** behind a port such as `Summarizer` or `Classifier`.
- The prompt template and output schema are the **contract** of that port — version them with the code and cover them with evals (`spec-driven.md` §4).
- The business decision made *with* the model's output ("if classified as fraud, hold the order") is **domain** or **app** code and is unit-tested with a fake adapter that returns canned outputs.
- Never let the domain import the provider SDK. Swapping providers should be an adapter change.

> **Sources:** OWASP [LLM Prompt Injection Prevention Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/LLM_Prompt_Injection_Prevention_Cheat_Sheet.html) — untrusted content and tool controls; [RAG Security Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/RAG_Security_Cheat_Sheet.html) — retrieval authorization, tenant/cache isolation, and fail-closed behavior.

## 6. Cross-References
- `../SKILL.md` — the decision protocol.
- `spec-driven.md` §4 — prompt specs and eval-driven development.
- `code-placement.md` — where the adapter, port, and business rule live.
- `engineering-principles.md` §14 — AI-assisted engineering hygiene.
