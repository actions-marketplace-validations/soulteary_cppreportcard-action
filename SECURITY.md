# Security

Treat analyzed repositories and build preparation as untrusted code. Run them in
isolated CI jobs with least-privilege tokens. Do not use pull_request_target to
checkout and build untrusted pull request code with privileged credentials.

The CLI does not execute compilation database command strings itself. It passes
build configuration to external analyzers; those tools remain part of the trust
boundary. No shell interpolation is used for analyzer invocations. The Action
commits reports only when explicitly enabled and only on branch push events.

Report suspected vulnerabilities privately through GitHub's repository security
reporting feature if available, rather than publishing exploit details in issues.
