# r/LocalLLaMA Draft

**Title:** I built an adversarial tester for LLM apps that runs 100% locally via Ollama

Hey everyone,

I've been working on a CLI tool called teststop. It's designed to act like a chaotic, adversarial user that actively tries to break your LLM agents and apps. 

A lot of the eval frameworks out there (Promptfoo, DeepEval, etc.) are built around static benchmarks, and often assume you're hitting an external API. I wanted a tool that dynamically generates and executes edge cases, but I didn't want to pay OpenAI or Anthropic a fortune just to test my own code. 

So, teststop defaults to using Ollama locally. It scans your project, generates adversarial scenarios, and then fires them at your running application (via HTTP) to see what breaks. Because it's local, you can run aggressive, high-concurrency test suites without worrying about token quotas or sending your application data to a cloud provider. 

It tracks pass/fail confidence in a `.teststop/memory.json` file over time, reducing tests on stable areas and focusing on the volatile parts.

It's completely free and open source. I'm currently trying to gauge if there's real demand for a purely local red-teaming tool before I build out the rest of the roadmap. 

Would love to hear if you find this useful or if you have any feedback on running heavy eval suites locally.

Repo: https://github.com/shaifulshabuj/teststop
