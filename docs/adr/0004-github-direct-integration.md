# Use GitHub Branches as the Code Integration Surface

AgentDock will integrate directly with GitHub for repository access, branch context, and code publishing.

GitHub is the source repository and code collaboration surface. AgentDock runs work in isolated sandboxes and owns issue, run, trace, patch review, permission, and sandbox lifecycle state.

**Decision**

AgentDock will use a GitHub App / connector model rather than asking sandboxes to hold long-lived user Git credentials.

The backend will:

- authenticate repository access through GitHub installation/user authorization;
- create one issue branch for each top-level issue;
- create runs from a selected branch and commit SHA;
- record the base commit used for each run;
- receive patches, logs, traces, and artifacts from sandbox execution;
- display patch diffs under the issue;
- apply accepted patches back to the corresponding GitHub branch according to the issue's permission policy.

AgentDock will not maintain a user-visible internal repository mirror or internal integration branch as a product concept. A temporary clone, checkout, or object cache may exist as an implementation detail, but GitHub branches are the code integration surface.

Sandboxes will not directly consume GitHub webhooks, own GitHub tokens, or push to protected branches as part of normal execution. The backend owns GitHub writes.

**Consequences**

GitHub is a first-class integration from the start, not a later export-only feature. Users can do pull request creation, review, and merge in GitHub instead of AgentDock implementing a full PR workflow in the MVP.

AgentDock still needs its own product state:

- issues and subissues;
- runs and run events;
- sandbox sessions;
- patch review state;
- patch versions;
- permission policy;
- branch and commit pointers;
- conflict and apply state.

GitHub branches hold accepted code changes. AgentDock holds the operational history around how those changes were produced: prompts, traces, logs, diffs, test output, permission decisions, and sandbox follow-up context.

Each top-level issue has exactly one issue branch in the MVP. Child issues inherit the parent issue branch instead of creating their own branches. This keeps a decomposed piece of work tied to one GitHub branch and one eventual pull request, while still allowing AgentDock to assign child issues as separate work items.

Multiple runs and child issues under the same top-level issue may produce multiple patch versions, but accepted code is serialized onto the shared issue branch.

The default permission level is manual. A completed run shows its patch diff under the issue and waits for the user to apply it. Higher permission levels may auto-apply a passing patch to the issue branch, but they still do not merge pull requests or push to protected branches.

Each active top-level issue may have one current sandbox session attached to it. Runs under that issue reuse the attached sandbox while it is alive. After a run finishes, the sandbox remains alive for the awaiting-review TTL so the user can continue work from the same environment. The default TTL is 60 minutes and can be extended by user action.

TTL expiry closes the sandbox but does not close the issue. If the issue branch is still active and later work is requested, AgentDock creates a new sandbox from durable issue state: the issue branch HEAD plus persisted issue context, patch history, traces, and comments. If needed, an unapplied latest patch version can be applied into the new sandbox as starting context.

While an attached sandbox is alive, users can choose Continue Work to resume from the exact sandbox context where the agent left off. This keeps the working tree, installed dependencies, transient files, and process-local context available for iterative follow-up. Continue Work produces a newer patch version for the same issue.

Issue branches use the MVP naming convention `agentdock/{issue_id}-{short_slug}`. AgentDock does not force-push issue branches in the MVP; every applied patch creates a normal commit on the issue branch.

This makes GitHub-side history readable and avoids rewriting branches that users may inspect or open pull requests from.

When applying a patch, AgentDock treats the latest GitHub issue branch HEAD as authoritative. The backend locks the issue branch, fetches the latest HEAD, and applies the patch against that commit. If the run's base commit is stale but the patch applies cleanly, AgentDock commits and pushes the result and records the patch as applied against a newer HEAD. If the patch does not apply cleanly, AgentDock records a conflict state and does not push.

Conflict resolution is explicit in the MVP. A conflicted patch can be rejected, retried from the latest branch HEAD, or sent back to an alive sandbox for follow-up work.

Agents should run a sync-before-apply workflow before producing an apply-ready patch. The workflow fetches the latest target branch, detects whether the branch moved, rebases or reapplies the agent's changes when possible, resolves conflicts when the agent can do so safely, runs checks, and emits a fresh patch version.

This workflow is an agent skill, not the source of truth. The backend still performs the final apply guard: lock the issue branch, fetch latest HEAD, apply the patch against latest HEAD, and only push if it applies cleanly. If another commit lands after the skill runs, the backend guard wins.

Patch review is versioned. A run may produce multiple patch versions while the issue's attached sandbox is alive. When the user requests follow-up changes, an alive sandbox continues the work and produces the next patch version. If the sandbox has expired, AgentDock starts a new run in a new sandbox from the issue branch HEAD and records the next patch version under the same issue.

Only the latest pending patch version can be applied. Older pending versions are marked superseded when a newer version is produced.

Run state and patch state are separate. Run state describes execution lifecycle; patch state describes code decision lifecycle. A completed run does not imply that its patch has been applied.

MVP run states are `queued`, `provisioning`, `preparing_workspace`, `running`, `awaiting_review`, `completed`, `failed`, and `cancelled`. `awaiting_review` is the interactive hold state where the sandbox may remain alive for follow-up changes.

MVP patch states are `pending`, `superseded`, `applied`, `rejected`, and `conflict`.

Sandbox retention follows both run TTL and issue branch lifecycle. For an active issue branch that has not been merged or closed, an alive sandbox can be reused so the user can continue where the agent left off. When the TTL expires, AgentDock closes the sandbox and can recreate a new one for later work on the same issue. Once the issue branch is merged, closed, or otherwise marked done, AgentDock should clean up any attached sandbox and preserve only durable records: branch pointers, commits, patch versions, traces, logs, and artifacts.

Cleanup is based on branch lifecycle, not exact commit hash equality. GitHub may merge by merge commit, squash, or rebase, so the final target-branch hash may differ from the issue branch HEAD even when the code has been accepted.

AgentDock determines merged state from an associated GitHub pull request, not from commit hash comparison. The normal path is GitHub `pull_request` webhooks matched by repository and head branch. A closed pull request with GitHub's merged flag marks the issue branch merged. A closed pull request without merged state marks the issue branch closed without merge.

Periodic reconciliation should verify webhook-derived state through the GitHub Pull Requests API. If a branch has no associated pull request, AgentDock must not mark it merged based only on commit ancestry or branch deletion. It remains active, closed, or unknown until the user or GitHub PR state provides a clear lifecycle signal.
