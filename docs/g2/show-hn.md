# Show HN Draft

## Candidate Titles
1. Show HN: An adversarial AI tester that acts like your most unpredictable user
2. Show HN: I built an AI eval tool that attacks your agent from the outside
3. Show HN: teststop – Autonomous red-teaming for your LLM apps

## Body
Hey HN,

I'm building teststop, a CLI tool that acts like a chaotic, adversarial user for your LLM applications and agents. 

Testing LLM apps is notoriously hard. We usually rely on unit tests, static benchmarks, or platforms like Promptfoo and DeepEval. Those are great for static assertions and CI/CD grading, but they suffer from "assumption coverage"—they only test the edge cases the developer already knew about. 

I wanted something that dynamically generates adversarial scenarios tailored to my specific codebase context and actually executes them against a running system. teststop does exactly that. It reads your codebase, figures out what you're trying to do, and then tries to break it.

It runs from the CLI and can execute scenarios against your live target via HTTP. By default, it uses a local Ollama model to generate and run these scenarios (so it's free and private), but you can point it to Claude if you want higher quality.

A heads up: because this is dynamically generating and executing adversarial scenarios, it can be slow and, if you use a cloud model, it costs tokens/time. It's also an early pivot from a previous project, so the edges are still rough.

I'm posting this to gauge demand before I build out the full MVP. If you're building LLM agents and struggling to find the cracks before your users do, I'd love your feedback. 

- Does the `.teststop/` config/cache directory approach make sense for this?
- Would you run something this slow/expensive in your pipeline?

Repo: https://github.com/shaifulshabuj/teststop
