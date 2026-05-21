# The Mandate

The mandate (`mandate/base.md`) is the most important file in the entire teststop project.

> The quality of this file determines the quality of every test scenario generated. Get this wrong and everything else fails regardless of how good the code is.

---

## What the Mandate Is

The mandate is the instruction that transforms a general-purpose AI into an adversarial user. It tells the AI:

1. Who it is (not a tester — a real user with real frustrations)
2. How real users actually behave (not how developers imagine they do)
3. What chaos conditions to consider
4. How to adapt its thinking to different system types
5. The exact JSON format to output

Without the mandate, AI generates generic, sanitized test cases that developers already thought of. With the mandate, AI generates the scenarios that production teaches you the hard way.

---

## The Core Problem the Mandate Solves

Here's the failure mode it addresses:

> **Developer mindset**: "A user fills in the form correctly and clicks submit."
>
> **Real user**: Fills in the form, goes to make coffee, comes back 40 minutes later, submits — and gets a session timeout error with no way to recover their form data.

AI knows both of these. But without explicit instruction, it defaults to the developer mindset. The mandate locks it into adversarial-user mode.

---

## Mandate Structure

The mandate has seven sections:

### 1. Identity Framing

Establishes who the AI is for this run. It is not a QA engineer. It is not a developer. It is a real user with real behavior patterns.

### 2. Adversarial User Behavior Patterns

Ten concrete patterns that real users exhibit:

| Pattern | Example |
|---------|---------|
| Double-submit | Clicking "Submit" twice because nothing seemed to happen |
| Stale sessions | Returning to a page hours or days later |
| Parallel tabs | Opening the same form in two browser tabs |
| Unexpected input | Pasting data from Excel with hidden characters |
| Network interruptions | Submitting a form on a flaky connection |
| Back-button abuse | Navigating back mid-workflow |
| Bookmark navigation | Jumping directly to a deep URL, skipping prerequisites |
| Concurrent users | Two users editing the same resource simultaneously |
| Bulk operations | Selecting all 5,000 items and clicking "Delete" |
| Permission probing | Trying URLs and actions they shouldn't have access to |

### 3. Chaos Conditions

Eleven environmental conditions that make failures more likely:

- Slow network (2G/3G, satellite, VPN)
- Browser quirks (autocompletion filling wrong fields, saved passwords injecting values)
- Mobile-specific (keyboard covering form fields, orientation changes mid-flow)
- Dependency failures (third-party services slow or down)
- Concurrent load (many users hitting the same endpoint simultaneously)

### 4. System Type Adaptations

The AI adjusts its thinking based on what kind of system it's testing:

| System Type | Focus Areas |
|------------|-------------|
| `web_app` | Form flows, navigation, session management, responsive behavior |
| `api` | Auth edge cases, rate limits, malformed payloads, concurrent requests |
| `cli` | Signal handling, partial input, pipe interruption, invalid flags |
| `data_pipeline` | Partial data, schema mismatches, reprocessing idempotency |
| `library` | Edge inputs, concurrency, version compatibility |
| `mobile_app` | Background/foreground transitions, low battery, poor connectivity |

### 5. Scenario Construction Guide

Instructions for how to build each scenario:
- Lead with the user's frustration, not the technical condition
- Include realistic preconditions (not "fresh session" — "user who signed up yesterday and just remembered their password")
- Specify concrete chaos factors, not vague ones

### 6. Context Injection Tokens

The mandate contains nine tokens that teststop replaces at runtime:

```
[SYSTEM_NAME]          ← project name
[DETECTED_LANGUAGE]    ← primary language
[DETECTED_TYPE]        ← system type
[DETECTED_FLOWS]       ← extracted routes and flows
[MEMORY_STABLE_AREAS]  ← proven areas (reduce coverage)
[MEMORY_VOLATILE_AREAS] ← unproven areas (increase coverage)
[N]                    ← number of scenarios to generate
```

### 7. JSON Output Contract

The mandate specifies the exact JSON schema the AI must return — 17 rules covering structure, field types, valid values, and formatting.

---

## Viewing the Mandate

See exactly what teststop sends to the AI:

```bash
teststop mandate --show
```

With project context injected:

```bash
teststop mandate --show --path ./my-project --depth aggressive
```

---

## Improving the Mandate

The mandate is the highest-leverage contribution to teststop.

**When to improve it:**
- You've seen a failure pattern in production that teststop didn't catch
- The AI consistently generates scenarios that are too generic
- A new system type needs better coverage

**How to improve it:**
1. Fork the repository
2. Edit `mandate/base.md`
3. Test with `teststop mandate --show --path <your-project>`
4. Run `teststop run` on several different project types
5. Compare scenario quality before/after
6. Submit a pull request with the improvement and the failure pattern that motivated it

**The bar for a good improvement:**
A real scenario, from a real failure, in production. Not theoretical. Not "what if". Something that actually broke.

---

## The Mandate Is Not a Config File

The mandate is universal. It does not need project-specific configuration — teststop injects the project context automatically. The mandate itself covers behavior that applies to all software systems regardless of what the system does.

Do not add project-specific behavior to the base mandate. That's what the injected `[DETECTED_FLOWS]` and `[DETECTED_TYPE]` tokens are for.

---

## Philosophy

> Every test ever written was written by someone who *knew* how the system works.
> Real users don't know.

The mandate exists to give the AI a state of productive ignorance — the same ignorance that real users have. It knows general patterns of user behavior. It does not know your system's assumptions.

This is why it works. The AI generates scenarios that your team never would — not because the AI is smarter, but because it wasn't there when your team built the assumptions into the code.
