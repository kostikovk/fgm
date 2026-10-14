# Developer Response Patterns

Choose the pattern that makes the current result easiest to assess. These are adaptable examples, not required headings or fixed templates. Required fields from the owning skill remain required.

## Contents

1. [Explanations and decisions](#explanations-and-decisions)
2. [Implementation reports](#implementation-reports)
3. [Progress and debugging](#progress-and-debugging)
4. [Code reviews](#code-reviews)
5. [Action plans](#action-plans)

## Explanations and decisions

Give the direct answer or recommendation, then the reason that decides it. Ground technical claims in the supplied artifacts, inspected code, or relevant sources. Separate what exists from what is proposed.

Use a small example to clarify a mechanism when prose alone is difficult to follow. For unfamiliar stack behavior, use [Cross-stack explanations](cross-stack-explanations.md). For alternatives, compare relevant trade-offs rather than listing every imaginable option. When missing information determines the answer, give a conditional recommendation or ask the specific question instead of presenting certainty.

Example using supplied project excerpts:

> Have the new request handler call `ProjectCreator.create()`. It already checks the account's project limit and saves the project, so the handler only needs to read the request and return the appropriate response.
>
> The supplied `project-creator.ts` and `project-quota.ts` excerpts show these responsibilities. They do not show a test with simultaneous requests against a real database.

Longer teaching requests can use sections and a worked example. Do not truncate the explanation merely to satisfy a brevity preference.

## Implementation reports

Lead with the observable change and why it matters. Group files by their contribution when there are many; mention the key paths rather than reproducing every diff. Keep actual verification beside the outcome, including checks that failed or could not run and any material remaining work.

Example with supplied completion evidence:

> Empty input now returns a validation error before database access.
>
> Updated `src/import.ts` and its regression test. The focused test and typecheck passed; the database integration suite was not run because its service was unavailable.

If the active TDD workflow requires the observed RED failure, GREEN results, and refactor decision, include them compactly. Do not replace that evidence with “tests passed.” Keep snippets for changes the developer needs to inspect; a linked file or diff is often sufficient for work already applied.

Finish after the result and verification when the task is complete. Include a remaining action only when it matters to completion, with who owns it.

## Progress and debugging

A progress update should identify what is known and what the next agent action will resolve. Restate context only when it helps after an interruption, a long gap, or a change of direction. Avoid narrating every file read or tool call.

For a failure, name the observed symptom and the evidence. Separate the leading hypothesis from the confirmed cause. Give the next useful diagnostic action without blame or an invented duration.

Example with supplied diagnostic evidence:

> The server receives the request but times out while waiting for the database. The trace shows the delay before the query starts; all database connections being busy is still a hypothesis. I’m checking available connections next.

When blocked, state what failed to run, the resulting verification gap, and the specific missing input or dependency. Perform safe, authorized recovery yourself before handing back an action. Do not ask the user to run an ordinary local command the agent can execute.

## Code reviews

Put actionable findings first, ranked by impact using the review workflow's severity convention. Each finding needs a location, a concrete trigger or scenario, and the resulting consequence. Put supporting detail immediately after the claim rather than burying it in a separate narrative.

Keep all material findings. If the list is long, group related findings or use brief summaries with detail below; a fixed item cap must not omit defects. Separate confirmed issues from unresolved questions and state review coverage limits.

Example based only on supplied code:

> `internal/projects/create.go`: two simultaneous requests can exceed the account's project limit because counting and saving are separate database calls. Both requests can see one remaining slot and then each save a project.

Do not invent severity labels, inspected line numbers, tests run, or a clean bill of health without the corresponding evidence. When no findings were identified, state the inspected scope and any relevant validation gap plainly.

## Action plans

Use a numbered sequence when order matters. Give each step a concrete action and the result that establishes it is complete. Put prerequisites before dependent commands. Keep commands copyable and use placeholders only when explicitly identified.

For branching procedures, show the condition beside the relevant branch instead of forcing a single misleading linear list. Provide enough steps to finish the requested operation; group a long procedure by phase when helpful.

Distinguish a plan from work already performed. If the agent is authorized and able to execute the plan, continue the task instead of ending with instructions for the user. Estimates are optional and require a stated basis when uncertainty matters.
