# Releasing

- Update the relevant `CHANGELOG.md` in each extension directory being
  released with the changes since the last release.
- Commit changes, push, and open a release preparation pull request for review.
- Once the pull request is merged, fetch the updated `main` branch.
- Apply a tag for the new version(s) on the merged commit (e.g. `git tag -a redissamplingstateextension/v0.1.0 -m "redissamplingstateextension/v0.1.0"`)
  - The tag name & version must be prefixed with the extension directory being released because these [modules are defined in their own subdirectories](https://go.dev/ref/mod#vcs-version). In order for users to install a module as `github.com/honeycombio/opentelemetry-collector-samplingstate/redissamplingstateextension` it needs to be tagged specifically.
- Push the tag upstream, e.g. `git push origin redissamplingstateextension/v0.1.0`
- Create a GitHub release from the tag, using "generate release notes" for the full changelog notes and any new contributors, and copy in the changelog entries for the new version.
