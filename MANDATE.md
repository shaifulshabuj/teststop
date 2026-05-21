# The teststop Mandate

> The mandate is the soul of teststop. Everything else serves the mandate.

---

## What the Mandate Is

The **mandate** is the instruction set given to an AI that makes it test software the way a real adversarial user would break it — not the way a developer who wrote the code would test it.

Every AI coding assistant already knows:
- How humans actually use software (and misuse it)
- What failure patterns occur in production across thousands of systems
- What edge cases emerge from real-world usage
- How to think adversarially about any system type

The mandate gives AI the *permission and framing* to apply this knowledge to your specific system.

---

## Why the Mandate is the Core

teststop without a great mandate is just a shell script.

The code that scans your project, scores confidence, and formats output — that's plumbing. The mandate is the brain.

A poor mandate produces:
- Developer-perspective tests ("the login button works when credentials are correct")
- Coverage of expected paths only
- Test suites that grow without reducing

A great mandate produces:
- User-perspective tests ("what happens when I refresh the page mid-checkout?")
- Adversarial scenarios developers never thought of
- A testing surface that shrinks as the system matures

**The quality of your test scenarios is directly proportional to the quality of the mandate.**

---

## How to Read the Mandate

```bash
teststop mandate --show
```

This prints the exact text sent to the AI — including the project context that teststop injected. What you see is what the AI receives.

The base mandate is at `mandate/base.md` in this repository. It is plain Markdown — human-readable, version-controlled, and community-editable.

---

## The Adversarial User Mindset

The mandate instructs AI to embody a user with these characteristics:

**They have never read your documentation.**
They do not know the "correct" way to use your system. They do what feels natural.

**They are not malicious, but they are unpredictable.**
They retry when something is slow. They navigate away and come back. They paste content from other apps. They open multiple tabs.

**They have real constraints.**
Slow or unreliable network. Old browser. Small screen. Accessibility needs.

**They do things in the wrong order.**
Developers test the happy path. Real users find the other 47 paths.

**They make mistakes and expect recovery.**
Wrong input. Accidental submission. Double-click on a button. The system must handle this gracefully.

---

## How the Mandate Evolves

The mandate improves through community observation of real failure patterns.

If teststop generates a scenario and it finds a real bug → the mandate works.
If teststop generates a scenario and it finds nothing → the mandate might need refinement.
If a real user breaks something that teststop never tested → the mandate needs updating.

**The mandate should improve with every production incident that testing missed.**

---

## Contributing to the Mandate

The mandate is the highest-leverage place to contribute to teststop.

### What makes a good mandate addition:

1. **Grounded in real failure patterns** — something that actually happens in production
2. **Universal** — applies to many system types, not just one specific app
3. **Framed as user behavior** — "what users do" not "what developers miss"
4. **Testable** — generates scenarios that can be evaluated

### How to contribute:

1. Fork this repository
2. Edit `mandate/base.md`
3. Run `teststop run` on a test project — evaluate whether the new scenarios are better
4. Open a PR with the mandate change and example scenarios it produced

### What not to add:

- Developer-perspective test patterns ("check that the API returns 200")
- System-specific logic that only applies to one type of app
- Anything that makes the mandate longer without making scenarios better

---

## Mandate Versioning

The mandate is versioned with the project. Breaking changes to the mandate schema are treated the same as breaking changes to the code API.

When the mandate improves significantly, `teststop` will regenerate fresh scenarios for volatile areas — because the AI now knows something it didn't before.

---

## The One Question

When evaluating any mandate change, ask:

> "Would this produce a scenario that a real user would actually encounter?"

If yes: it belongs in the mandate.
If no: cut it.

---

*The mandate is the product. Everything else is infrastructure.*
