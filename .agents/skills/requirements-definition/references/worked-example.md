# Worked example: team invitations

Use this example to see how a loose request becomes a scoped contract. All product rules and decisions below are illustrative; they are not defaults for real invitation systems.

## Incoming request and intake

Request: “Let administrators invite people to their team so support no longer has to add every member.”

The intended improvement is administrator self-service. The proposed invitation mechanism is part of the request; details such as account eligibility, identity matching and expiry remain unspecified. A code inspection might reveal existing manual membership creation, but that would establish current behavior rather than the desired invitation policy.

Questions that affect delivery include: may invitees lack an account, who can accept, what role is granted, and what makes an invitation invalid? Product outcome evidence also needs a plan: successful invitation acceptance is observable immediately; reduction in support work requires later observation and an agreed comparison.

## Illustrative decisions and selected scope

For this example only, assume the requester explicitly resolves these decisions: the first delivery supports existing accounts; only the addressed account accepts; new members receive Member access; invitations are single-use and expire after seven days. A later decision about resending or revocation is deferred. These are confirmed illustrative decisions, not assumptions silently chosen by an agent.

| ID | Intended behavior | Source / rationale | Status | Acceptance examples / evidence plan |
|---|---|---|---|---|
| R-INV-1 | An authorized team administrator can create an invitation for an existing account | Illustrative requester decision; self-service membership | Confirmed for first scope | Administrator creates an invitation; regular member is refused without creating one; API/UI acceptance |
| R-INV-2 | Only the addressed account can accept a valid invitation into the stated team | Illustrative identity decision; prevent unintended access | Confirmed | Addressed account gains Member access; different account is refused; identity and scenario checks |
| R-INV-3 | Expired or used invitations cannot create membership; successful acceptance consumes the invitation without duplicate membership | Illustrative lifecycle decision; protect membership integrity | Confirmed | Expired/used invitation changes nothing; repeated or concurrent acceptance creates at most one membership; integration checks |

Unregistered invitees, resending and revocation remain visible as deferred scope. They have no delivery evidence. Existing contracts concerning team limits or duplicate members must be inspected before declaring the example ready in an actual repository; this example does not assume them away.

## Ready-scope handoff

Hand off one candidate scenario: an administrator invites an existing account, the addressed user accepts, and the user gains access to the team. Reference R-INV-1 through R-INV-3 and the illustrative decisions. task-decomposition determines coherent technical tasks and dependencies; requirements-definition does not prescribe a database or endpoint layout.

If the account-eligibility decision had remained open, mark the invitation scope blocked rather than inventing registration behavior. Unrelated settled work can proceed only when its contract does not depend on that decision.

Technical completion requires the relevant task checks. Assembled acceptance requires the entire invitation-to-access path and refusals to work together. Product validation additionally checks whether administrators can accomplish the intended self-service need; any support-work reduction remains pending until observed.

## Later change

Suppose the requester changes expiry to 24 hours. Update R-INV-3 and the contract history; trace creation, acceptance, time-dependent examples, displayed expiry and existing evidence. Mark affected plans/evidence stale and reassess them against the new rule. Evidence for administrator permissions can remain usable if the change demonstrably leaves that behavior unaffected. New-account support is still deferred unless separately requested.
