# Include Eval in the MVP

AgentDock will include eval as an MVP product capability because evaluating agent work is central to the platform, not an optional analytics layer. Checks answer whether repository commands passed; eval answers whether the agent's work was good, complete, safe, and useful.

The MVP eval surface should stay lightweight: persist eval inputs and outputs for each run, show them on the issue/run view, and support a small set of built-in evaluators before adding external eval SDK integrations. Eval should consume durable records such as the issue prompt, comments, patch version, trace, check results, apply/reject outcome, and user feedback.

External eval SDKs can be added later behind an evaluator interface. The MVP should avoid making external eval infrastructure a prerequisite for running an agent.
