# r/LangChain Draft

**Title:** An external adversarial tester for your agents (complements LangSmith)

Hey r/LangChain,

If you're building complex agents, you're probably already using LangSmith or Braintrust for observability and trace evaluation. Those tools are fantastic for deeply integrated, white-box testing.

I'm building teststop, a CLI tool that takes the opposite approach: it acts as a completely framework-agnostic, external attacker. 

Instead of asserting on the internal state of your chain, teststop dynamically generates adversarial, chaotic user scenarios based on your codebase and fires them at your running application via HTTP. It's designed to find the emergent, unpredicted failures that your static tests and standard eval datasets miss—like prompt injections, weird user inputs, or race conditions.

You can think of it as a synthetic chaotic user that generates the weird traffic your observability platform (like LangSmith) then monitors. 

It runs entirely from the CLI and uses a local Ollama model by default (or Claude if you prefer). It outputs standard exit codes (0–3) and JSON, so you can easily drop it into your CI/CD pipeline alongside your existing LangChain evals.

I'm in the early stages of pivoting this project toward general LLM evaluation and want to see if other agent builders find this "external attacker" angle useful. Let me know what you think!

Repo: https://github.com/shaifulshabuj/teststop
