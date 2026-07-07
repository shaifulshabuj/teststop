# FAQ Replies

**How is this different from promptfoo/deepeval?**
Promptfoo and DeepEval are excellent for static benchmarks, unit testing, and matrix evaluations against known edge cases. *teststop* is different—it dynamically generates adversarial scenarios based on your specific application context and executes them externally. It's meant to autonomously discover the *unknown* cracks in your logic, acting more like an unpredictable red-team user than a static grader.

**Does it work with my stack / no LangChain?**
Yes, it's 100% framework-agnostic. teststop attacks your system externally via HTTP or CLI. As long as your application is running and accessible (e.g., `localhost:8080`), teststop can test it, regardless of whether you wrote it in LangChain, raw Python, Go, or anything else.

**What does a run cost / how slow is it?**
Because teststop is dynamically generating and executing AI-driven scenarios, it is inherently slower and more expensive than static tests. If you use the default local Ollama model, it only costs you time. If you configure it to use Claude, you'll be paying for the tokens consumed during the scenario generation and evaluation phases. Expect runs to take minutes rather than milliseconds.

**Is the output deterministic / CI-usable?**
Yes. While the scenario generation is AI-driven, the execution outcome is strict. teststop outputs standard exit codes: `0` (Safe to deploy), `1` (Review required), `2` (Critical failures), and `3` (Internal error). It also supports `--output json` with a frozen schema for easy integration into your CI/CD pipelines.

**Why the `.teststop/` dotdir?**
teststop needs a place to store its confidence memory (`memory.json`) and configuration so it can track which parts of your app are stable over time and focus its attacks on volatile areas. It's similar to `.pytest_cache` or `.next`. I'm open to adding a global override if people hate dirtying their repo, but checking the memory file into version control is actually recommended to prove system stability.

**What models does it need?**
By default, it requires a local installation of Ollama (it looks for `qwen3.6:latest` or similar). This keeps everything free and private. You can also configure it to use the `claude` or `copilot` CLI tools if you want higher-quality, cloud-based evaluations.

**Is this abandoned / what's the roadmap?**
It's not abandoned, but it is in an early pivot stage. I'm currently using this to gauge actual developer demand before committing to building out the full general MVP. If the community finds value in this autonomous adversarial approach, the roadmap will focus on better target integrations and smarter, faster scenario generation.
