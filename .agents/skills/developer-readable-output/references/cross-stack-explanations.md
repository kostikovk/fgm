# Cross-Stack Explanations

Assume general development knowledge, not expertise in every framework or backend mechanism. Use the user's stated background when available; do not label someone a beginner because they work in another stack. Explain enough to assess the change without forcing the reader to learn the entire subsystem.

## Contents

1. [Explain the mechanism](#explain-the-mechanism)
2. [Make the reasoning checkable](#make-the-reasoning-checkable)
3. [Worked examples](#worked-examples)

## Explain the mechanism

Start with the behavior in familiar terms. Keep a specialist term when it is useful for locating code, documentation, or a design, and explain it where it first matters. For example, say “both requests can see a free slot before either saves its project” before naming a race condition.

Explain acronyms through their role in this change rather than expanding letters alone. “We used an outbox” does not explain when an event is saved or sent. “We applied SOLID” does not identify a responsibility or show why the new boundary helps. Use the actual mechanism supported by the artifacts.

Prefer a short input → behavior → result example when sequencing or state is hard to understand. Preserve distinctions that affect correctness: waiting is not rejection, a retry is not guaranteed success, and a passing test covers its scenario rather than proving every execution safe. Do not invent an implementation detail to make the story simpler.

## Make the reasoning checkable

For a meaningful change, connect the observed problem, the chosen mechanism, and its practical effect. Include the reason for the choice when known; if the artifacts establish only what changed, do not invent the original author's motivation.

Use concrete support: an identified function and its behavior, a relevant request example, a log observation, or the scenario and outcome of an executed test. A file path tells the reader where to look; explain what in that file supports the claim. A pattern name or a claim that something is modern is not evidence.

Discuss an alternative only when it explains the actual decision. For example, if two requests can pass a separate count check, explain why moving that check to a helper alone would leave the same problem. Do not fabricate a comparison or benchmark that was not performed.

Keep supplied excerpts, repository inspection, tests actually run, and proposed behavior distinct. State relevant coverage gaps in plain language. Avoid universal claims such as “duplicates are impossible” when only one path or scenario was verified.

## Worked examples

These examples use synthetic supplied evidence. Their paths and test results are illustrative, not repository findings.

### Backend change for a frontend developer

Supplied evidence: `createProject` previously counted and inserted projects in separate database calls. Two requests could both see four projects under a five-project limit. It now holds an account-row lock while counting and saving; the concurrent-request database test passed. The import path was not inspected.

> Project creation now makes requests for the same account wait their turn while checking the limit and saving.
>
> Previously, two requests could both see four projects and each create another, leaving six under a limit of five. `createProject` now locks the account's database row during the check and save, so the second request checks the updated count after the first finishes.
>
> The database test with two simultaneous requests passed. The import path was not inspected, so this does not establish that imports follow the same rule.

### Frontend change for a backend developer

Supplied evidence: `SearchResults` applied every completed response. The request for “ca” could finish after the request for “cat” and overwrite its results. It now assigns a sequence number to each search and applies only the latest search's response; the delayed-response test passed. Requests are not cancelled.

> Search results now stay tied to the latest text the user typed.
>
> An older request could finish late and replace the results for a newer search. `SearchResults` now tracks which search was started most recently and ignores responses from earlier searches. The delayed-response test passed.
>
> Earlier requests still finish; this change prevents them from replacing the displayed results.
