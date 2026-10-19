# Ship session memory: 053-S (061-F) test and script hygiene follow-ups

* Mode: P-017 dark mode; merge pre-authorized; admin fallback NOT authorized; cascade-close NOT pre-authorized.
* Branch: chore/test-and-script-hygiene-follow-ups (first attempt used a non-matching slug; renamed before any commit; post_claim topology gate then passed).
* Tasks 061.001-T..061.006-T done; code commits d146c11..3fd3e94 plus remediation a475eaf.
* Local review: 3-reviewer parallel adversarial review (correctness, scope, Go safety) found one HIGH-consensus P1 (non-Windows vacuity assertion) and P3s; fixed in a475eaf; re-review READY.
* Out-of-scope findings captured per P-021 C2 as stash entries: D67ACCC7, 8CC8D27D (vacuous start_script tests), AD132ACF (flaky wrapper test).
* UNC junction test is SKIPPED on this host (Windows will not resolve UNC-target junctions); substitute-name form covered by TestJunctionNames.
* Full `go test ./...` passed with Git Bash first on PATH; default PATH has WSL bash first so bash wrapper tests skip with an explicit reason.
* Next: PR, Copilot review loop, merge commit, post-merge closure.
