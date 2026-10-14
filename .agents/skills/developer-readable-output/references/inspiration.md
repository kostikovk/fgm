# Inspiration and Scope

The user selected [ayghri/i-have-adhd](https://github.com/ayghri/i-have-adhd) as inspiration for readable coding-assistant output. Its [README](https://github.com/ayghri/i-have-adhd/blob/main/README.md) and [skill entrypoint](https://github.com/ayghri/i-have-adhd/blob/main/skills/i-have-adhd/SKILL.md) emphasize making the useful answer easy to find, structuring actions, avoiding tangents, and describing errors plainly. The current entrypoint also allows exceptions for completeness and task requirements.

This skill uses independently written guidance for developer communication. It is not a vendored copy or an installer for the upstream plugin. The upstream repository publishes an [MIT license](https://github.com/ayghri/i-have-adhd/blob/main/LICENSE); retaining copied or substantially adapted upstream text would require its copyright and permission notice.

## Adaptation choices

- Lead with the answer, result, or blocker appropriate to the request. An instruction belongs first when the developer needs to act.
- End with a concrete next action when work remains; completion does not require another task or a closing question.
- Keep responses scannable without limiting the number of material findings or required workflow evidence.
- Give timing only when supported and useful. State observed failures separately from inferred causes.
- Keep language, detail, and exact output contracts under the user's control. Presentation preferences do not imply a diagnosis or a claim of clinical effectiveness.

Upstream links follow `main` and may change. This reference records inspiration and adaptation rationale, not a requirement to fetch or install upstream on each invocation.
