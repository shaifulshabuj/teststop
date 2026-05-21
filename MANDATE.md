# The Mandate

The mandate is the intellectual heart of teststop.

If everything else were thrown away, the mandate alone — given to any capable AI — would still deliver most of the value of this project. The Go code in this repository exists to deliver the mandate well: to assemble it with relevant context, send it to a model, and turn the response into something a developer or another agent can act on.

This file explains what the mandate is, why it matters, and how to improve it.

---

## What the mandate is

The mandate is a single instruction document — currently `mandate/base.md` — that the AI receives when teststop runs. It tells the model exactly one thing:

> Test this system the way a real adversarial user would, not the way a developer would.

It is plain Markdown. It is open. It is embedded in the binary verbatim so anyone can audit it with:

```bash
teststop mandate --show
```

There is no hidden second prompt. There is no secret system message. The mandate is the whole instruction.

---

## Why it is the core

Every test suite ever written was written by someone who already knew how the system works. That is the source of the **assumption coverage problem**: we test what we believe will happen, not what actually happens.

Modern AI already knows how real humans behave with software. It has seen the retries, the abandonments, the pasted credit cards with hyphens, the two-tab races. It just needed an instruction that gives it permission to think like a user instead of like a developer.

The mandate is that instruction.

The Reader gives the mandate context. The Memory layer tells the mandate what is already proven. The AI Adapter ships the mandate to a model. The Reporter turns the model's response into something other systems can use.

The mandate itself is what makes the output reality-based instead of assumption-based.

---

## What makes a good mandate

A mandate change is good when it:

1. **Moves the model toward the user's perspective and away from the developer's.** Adding "what does the user feel when this is slow" beats adding "what does the API contract say."
2. **Names concrete behaviors instead of abstract ones.** "Pastes a credit card with spaces" beats "tests input validation."
3. **Stays language-agnostic and system-type-agnostic.** The mandate must work on a Next.js app, a Go CLI, and a COBOL batch job.
4. **Forces a mix of scenario classes.** If a small change skews the model toward only one class (only happy paths, only chaos, only adversarial), it is a regression.
5. **Keeps the output contract intact.** The JSON schema in the mandate must stay in sync with `pkg/scenario/types.go`. If you change one, you change both, in the same commit, and you bump the contract version.

A mandate change is bad when it:

- Adds developer-flavored hints ("check for SQL injection here") that nudge the model back into code-review mode.
- Adds project-specific knowledge that breaks universality.
- Pads instructions for cleverness rather than clarity.
- Tries to fix a single model's quirks at the cost of working on every other model.

---

## How to improve it

1. Read the current `mandate/base.md` end-to-end.
2. Run teststop against three different projects: a small one you know well, a medium-sized one you don't, and something exotic (a 20-year-old codebase, or a language you've never touched).
3. Read the scenarios the AI generated. Mark each as: *useful*, *generic*, *developer-flavored*, or *missed the user*.
4. Find the smallest change to the mandate that would move generic and developer-flavored output toward useful.
5. Run it again on the same three projects. Compare.
6. Open a PR that includes:
   - The mandate diff
   - Before/after scenario samples (anonymize as needed)
   - A short note on what category of scenarios improved and which got worse, if any
7. Bump the mandate version in the file header when the change is non-trivial.

---

## Versioning the mandate

The mandate is versioned along with the binary. The output contract — `pkg/scenario/types.go` and the JSON block inside `mandate/base.md` — is versioned more strictly:

- Adding a new optional field is a **minor** change.
- Renaming, removing, or changing the meaning of an existing field is a **breaking** change and requires bumping the major version of teststop.

When you change the mandate, do it in the canonical file at `mandate/base.md`. It is embedded into the binary at build time via `go:embed`; there is no second copy to keep in sync.

---

## The contributor's compass

When in doubt about a proposed change to the mandate, ask:

> *Does this make the model behave more like a real user — or more like a developer reviewing code?*

If it nudges the model toward "developer reviewing code", do not merge it, no matter how clever it is.

The mandate is the contract between teststop and reality. Treat it that way.
