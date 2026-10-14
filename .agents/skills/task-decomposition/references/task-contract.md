# Task contracts and coherence

Read this before handing tasks to an executor or checking whether a backlog can satisfy its selected product scope. Requirement meaning, source status, product ambiguity, and requirement changes belong to [requirements-definition](../../requirements-definition/SKILL.md) and its [requirement contract](../../requirements-definition/references/requirement-contract.md#requirement-record).

## Execution task contract

Use the project's existing task format. Fields can be compact prose for a small task; add detail only where it changes execution or acceptance.

| Field | Required meaning |
|---|---|
| Identity and traceability | Stable task identifier, parent slice/outcome, and canonical requirement references; include the applicable requirement revision or decision baseline |
| Result and kind | Observable contribution; distinguish implementation, enabler, and research spike |
| Scope | Included behavior, affected boundary, and relevant exclusions; do not silently add adjacent work |
| Context and authority | Canonical requirement/example/decision links, verified relevant implementation and contract locations, known constraints, and who may resolve outstanding decisions; record whether the request authorizes planning or implementation |
| Dependencies and contracts | Required inputs, predecessor results, producer/consumer responsibilities, interfaces or invariants needed for integration; link established contracts or flag missing decisions |
| Acceptance checks | Conditions that distinguish completion from failure, including relevant boundaries and regressions; match checks to the task kind |
| Evidence | Expected checks, observed outputs or artifacts, and limits of their proof; do not claim unexecuted checks passed |
| Status and stop conditions | Ready, provisional, blocked, or deferred with the reason; finish when acceptance evidence is collected, or return to requirements/design when a deciding assumption fails |

Reference relevant sources rather than copying a full product brief into each task. Provide enough context to execute the bounded work without relying on forgotten chat history. If the contract source is unavailable, identify the gap rather than inventing its contents.

An enabler's acceptance establishes the capability consumed by later work, such as a verified migration or an agreed reusable interface. A spike's acceptance is evidence answering its question, including uncertainty that remains; code produced during investigation is not automatically a production deliverable. Implementation follows [code-change](../../code-change/SKILL.md), including its testing workflow. Open architecture and API specification choices remain owned by [design-decision](../../design-decision/SKILL.md).

## Readiness versus provisional planning

Distinguish **contract-ready** from **ready to execute**. A contract is ready when required behavior, verified relevant repository context, constraints, dependency contracts, authorized scope, and verification path are sufficiently defined for its kind. The task is ready to execute when its start prerequisites are also satisfied. A pending predecessor can leave the contract ready while blocking execution; record both facts and schedule the task after the prerequisite is satisfied. Follow the project's status labels but make this distinction explicit.

Mark **provisional** when a candidate task is useful for planning but a deciding requirement, technical choice, or repository fact remains unresolved. Mark **blocked** when it cannot proceed until a specified input or prerequisite is supplied. Mark **deferred** when it is outside the selected delivery. Status is not permission: a contract-ready task in a planning request does not authorize execution.

Do not require all future scope to be known before beginning independent ready work. Conversely, a list of titles is not an executable contract, and an AI-proposed answer is not evidence that a product decision was made.

## Coverage and dependency review

Create a compact mapping or table where scope is large enough to hide gaps: selected requirement → slice → contributing tasks → acceptance evidence. Allow many-to-many links and mark partial coverage explicitly.

Review it for these failure modes:

- **Missing coverage:** a selected requirement or important acceptance example has no task and no assembled check. Add the work or make a deliberate scope decision through requirements-definition.
- **Unjustified tasks:** a task supports neither a selected requirement nor a necessary prerequisite. Remove it, defer it, or establish its actual justification.
- **Duplicate or absent responsibility:** several tasks assume another implements the same boundary, or no task owns it. Assign a responsible contribution and an integration acceptance responsibility.
- **Incompatible contracts:** producers and consumers disagree about inputs, outputs, errors, state ownership, access, or required guarantees. Reconcile established requirements and route open technical contracts to design-decision before affected execution.
- **Cycles or ordering errors:** record prerequisites as a directed dependency graph and check for cycles. A cycle may indicate that coupled tasks should merge or that a contract/enabler must be established first; do not hide it with arbitrary ordering.
- **Hidden operational prerequisites:** a slice relies on unverified delivery, migration, configuration, permissions, or external capabilities. Identify evidence and authority needed; local task planning does not grant external authority.

Shared requirements such as access control need coverage at every affected boundary and in assembled acceptance where relevant. One task label does not prove all consumers satisfy the guarantee. Parallel work is viable only where dependencies and contracts permit it; no particular agent arrangement is required.

## Assign acceptance and preserve evidence boundaries

Use the canonical distinction between task verification, assembled-scenario acceptance, and product outcome evidence in [Evidence and product outcomes](../../requirements-definition/references/requirement-contract.md#evidence-and-product-outcomes). Task preparation must identify how each relevant claim will be assessed; success at one level does not establish the others.

Assign an assembled acceptance responsibility before execution. Define a representative end-to-end path and relevant guarantees, using the project's actual validation environment. Isolated unit checks and mocked downstream behavior are useful task evidence but do not prove assembled behavior. If the real boundary cannot be exercised, report the remaining acceptance gap and the required environment or input; do not claim the slice is accepted.

Product outcome evidence may arrive after delivery. Preserve its owner, intended measure or observation, and timing where defined; report pending outcome evidence separately from completed engineering checks. Do not invent numeric success targets.

For requirement changes, apply [Requirement changes](../../requirements-definition/references/requirement-contract.md#requirement-changes). Use this decomposition's requirement-to-task mapping and dependency graph to supply its impact analysis, including producer/consumer contracts and assembled checks.
