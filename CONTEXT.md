# AgentDock Context

AgentDock is a managed platform for assigning coding work to AI agents and executing that work in remote sandboxes. This glossary keeps product language precise while the PRD and architecture are being shaped.

## Language

**Work Management Shell**:
The user-facing surface for organizing work around agents: workspace, repository, agent profile, issue, run detail, and patch review.
_Avoid_: full collaboration suite, Linear clone

**Remote Run Engine**:
The execution core that provisions a remote sandbox, prepares a repository, runs a coding agent, records trace events, and returns a result.
_Avoid_: local daemon, generic runtime

**Issue**:
A user-visible work item that describes coding work to be assigned to an agent.
_Avoid_: ticket, task request

**Run**:
One execution attempt for an issue by an agent in a prepared remote environment.
_Avoid_: task, job

**Run State**:
The execution lifecycle state of a run.
_Avoid_: patch status, review status

**Agent**:
A configured AI worker identity that can be assigned an issue and produce a run.
_Avoid_: model, provider

**Sandbox Session**:
The remote compute environment currently attached to an active top-level issue, including its repository checkout, process state, and cleanup policy. It may be reused by later runs while alive, but can expire after its TTL and be recreated from durable issue state.
_Avoid_: runtime, daemon, machine

**Continue Work**:
A user action that resumes agent work from an alive sandbox session after a patch version has been produced.
_Avoid_: rerun, restart

**Run Event**:
An append-only record of something that happened during a run, used for live viewing and later replay.
_Avoid_: log line, transcript message

**Patch Review**:
The issue-level review state for code changes produced by a run, including the diff, summary, checks, and the decision to apply, reject, or request follow-up.
_Avoid_: pull request, merge request

**Patch Version**:
One concrete diff produced during patch review for an issue. Newer versions supersede older pending versions.
_Avoid_: draft, revision

**Patch State**:
The decision lifecycle state of a patch version.
_Avoid_: run status, job status

**Check**:
A deterministic command result from the repository, such as tests, lint, typecheck, or build.
_Avoid_: eval

**Eval**:
A product-level assessment of agent work quality for an issue or run, using persisted inputs such as prompt, patch, trace, checks, and review outcome.
_Avoid_: test, lint

**Issue Branch**:
The GitHub branch that receives accepted code changes for a top-level issue. Child issues inherit the parent issue branch.
_Avoid_: internal branch, mirror branch

**Issue Branch State**:
The lifecycle state of an issue branch as known through GitHub branch and pull request events.
_Avoid_: merge status, git status
