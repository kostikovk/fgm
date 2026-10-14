# Worked example: a complete invitation scenario

This illustrates mapping, not a product specification to reuse. The decisions and IDs below are fictional and explicitly agreed for the example. Real work must inspect the actual repository and establish its own requirements through [requirements-definition](../../requirements-definition/SKILL.md).

## Selected product scope

Product request: “Administrators should invite people to a team.”

Illustrative first-delivery choices: existing registered users only, an identified recipient, Member role, a defined expiry policy, one successful use, and no bulk invitations or resend/revoke flows yet. No expiry duration, stack, delivery mechanism, or repository location is implied.

| Example requirement | Agreed behavior |
|---|---|
| R1 | Only an authorized team administrator can create an invitation for that team |
| R2 | The intended registered recipient receives a usable invitation through the agreed delivery channel |
| R3 | The authenticated intended recipient can accept a valid invitation and obtain Member access to the specified team |
| R4 | Expired, wrong-recipient, or already-used invitations do not create or alter membership; acceptance preserves single-use behavior, including concurrent attempts |

The complete slice S1 is **invite an existing user, deliver the invitation, accept it, and obtain access to the intended team**. Creating a record alone is a technical contribution. Expiry rules and the delivery channel must be resolved before their dependent tasks become ready; unrelated ready discovery can continue.

## Contributing work

These candidate boundaries require confirmation against the real code. They do not mandate separate tickets if the work is more coherent together.

| Task | Contribution | Requirement links | Dependency and evidence |
|---|---|---|---|
| T1 Create invitation | Persist the identified recipient and team with the agreed validity and authorization behavior | R1, R3, R4 | Agreed lifecycle contract; creation/access/boundary checks |
| T2 Deliver invitation | Send a usable acceptance route through the agreed channel | R2, R3 | T1 producer contract and established delivery capability; delivery/recipient checks |
| T3 Accept invitation | Validate recipient and invitation; establish membership and consume the invitation with the required integrity guarantee | R3, R4 | T1 lifecycle contract and existing membership behavior; valid/expired/wrong-recipient/reuse/concurrency checks |
| T4 Provide interaction | Let the administrator initiate and recipient complete the scenario with agreed observable outcomes | R1, R2, R3, R4 | Creation/acceptance/delivery contracts; interaction and error checks |
| T5 Accept assembled S1 | Exercise creation, real delivery boundary in the available test environment, acceptance, and resulting team access | R1–R4 | T1–T4 integrated; assembled success and relevant rejection evidence |

T1 and T3 share R4 responsibility; T4's error presentation does not establish the underlying integrity guarantee. Tests belong in the tasks whose behavior they verify. T5 supplies additional assembled evidence rather than postponing all testing until integration.

If the existing delivery capability is unknown, add a spike with a specific question, such as whether the existing channel can deliver the required acceptance route in the available environment. Its output is capability evidence and limitations. A necessary new delivery capability may become an enabler only after that evidence and applicable decisions establish its scope.

## Example execution contract for T3

- **Result:** the intended authenticated recipient can accept a valid invitation once and obtain Member access to its team.
- **Traceability:** S1; R3 and R4 at the agreed requirement baseline.
- **Scope:** acceptance and required membership/invitation integrity. Creation, delivery, bulk invitation, resend, and revocation are outside this task.
- **Context and authority:** canonical examples and lifecycle decisions; verified acceptance, identity, and membership locations after repository inspection. Product decision owners and authorization to implement must be recorded for the real request.
- **Dependencies:** T1's agreed invitation contract and verified membership behavior. Open transaction or concurrency choices go to design-decision; this example prescribes the guarantee, not an algorithm.
- **Acceptance:** valid intended-recipient acceptance establishes the specified membership; wrong-recipient, expired and repeated use fail without membership changes; concurrent attempts preserve the required single-use behavior and avoid duplicate membership.
- **Evidence:** relevant automated behavior and regression checks, plus actual outputs. Dependency substitutes do not prove delivery or the assembled scenario.
- **Stop conditions:** finish after contract verification; return for clarification if expiry, identity matching, existing-member semantics, or another deciding behavior is unresolved. Do not select a product policy implicitly.

Later slices for unregistered users, revocation, resending, or bulk invitations remain coarse and explicitly deferred. They are not required to make T3 ready if they do not affect its agreed behavior.

## Accept the result at the right level

Task checks establish their bounded contributions. S1 acceptance establishes that an administrator can invite the intended person and that this person obtains access through the assembled workflow, while relevant unauthorized or invalid attempts fail safely. If delivery is mocked, record delivery and assembled acceptance as unproven.

Whether administrators actually onboard teammates more effectively requires product outcome evidence agreed with the product owner. Successful engineering checks do not demonstrate that outcome, and this example supplies no invented adoption metric.
