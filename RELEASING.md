# Releasing

Once all the changes for a release have been merged to master, ensure the following:

- [ ] tests are passing
- [ ] user facing documentation has been updated

## Publish Release

To make a release, do these steps
1. Run `make ver=X.Y.Z version`
1. Update CHANGELOG.md
1. Merge changes
1. Run `make release_tag`