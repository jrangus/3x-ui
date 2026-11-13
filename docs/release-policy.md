# jrangus release and update policy

- `dev-6d205a91-backup` is the immutable backup of the original stable dev build.
- `3x-ui-YYYYMMDD` tags and their assets are immutable dated releases.
- `dev-latest` is only a compatibility pointer. It advertises one dated release; installers resolve that identity and then download every file from the dated tag.
- `jrangus-stable` is the canonical source branch for the next custom release. Upstream synchronization on `main` must not be used as a release source.
- A release build must inject the dated tag into the binary, publish SHA-256 sidecars, and pass the full build matrix before `dev-latest` is updated.
- Linux and Windows build jobs must execute the newly built binary and require its reported version to equal the requested dated release before packaging.
- The updater rejects a binary whose reported version differs from the resolved dated release and never combines binaries, scripts, or service files from different refs.

Rollback uses an explicit older dated tag. The update channel and previous dated releases are never overwritten to perform a rollback.
