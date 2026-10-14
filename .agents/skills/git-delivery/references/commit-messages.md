# Commit Messages

Use this reference for message suggestions or while preparing a requested commit. A message-only request does not authorize staging or committing.

## Discover the convention and change

Read applicable project instructions, contribution guidance, commit templates, relevant commit history, and existing commitlint/release configuration. Distinguish enforced rules from preferences. In a new repository, the lack of history is expected. Honor the project's actual format, language, scopes, issue references, and tooling.

Inspect the intended change or use the supplied diff/description, labeling incomplete evidence. Describe what will actually be committed, not unrelated unstaged changes or the entire conversation. For a local commit, [review its boundary and index](local-commits.md#choose-a-coherent-commit-boundary) before finalizing the message.

## Conventional Commit fallback

Without an established convention, use an English Conventional Commit:

```text
<type>[optional scope][optional !]: <description>

[optional body]

[optional footers]
```

- **`feat`** adds a capability; **`fix`** corrects a defect. Choose the type from the resulting behavior, not the file extension.
- Use other types such as `docs`, `refactor`, `test`, `build`, `ci`, or `chore` when they fit the change and the project accepts them. They do not inherently imply a semantic-version bump. `refactor` preserves behavior; a defect correction is still a `fix`.
- Use a scope only when it identifies a meaningful affected area. Prefer existing scope names; omit it when the change has no useful single scope.
- Describe the concrete outcome with a concise imperative phrase, such as `feat(skills): add Git delivery workflow`. Imperative wording, English, and subject-length limits are local preferences, not requirements of the Conventional Commits specification. Follow enforced project limits without introducing an arbitrary universal limit.
- Mark a breaking contract change with `!` before the colon or a `BREAKING CHANGE:` footer, regardless of type. If only `!` is used, the description must explain the breaking change. Explain migration when needed. Do not mark a change as breaking solely because its diff is large. Check the project's release tooling before relying on version automation.
- Use a body when the reason, observable before/after behavior, or migration is not clear from the subject. Include issue footers only when an actual supplied issue and the project's linking conventions justify them; do not invent references or verification results.

## Pass the message safely

For a multiline message, prefer a message file passed to `git commit -F`. Preserve actual newlines and literal content rather than interpolating a description into shell code. Use this command only as part of an authorized, verified local commit, not a message suggestion.

## Sources

[Conventional Commits 1.0.0](https://www.conventionalcommits.org/en/v1.0.0/) defines the message format and breaking-change markers. [Git commit](https://git-scm.com/docs/git-commit) documents message files and templates. The English default and imperative wording encode the user's workflow preferences.
