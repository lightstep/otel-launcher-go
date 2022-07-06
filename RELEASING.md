# Releasing

Once all the changes for a release have been merged to master, ensure the following:

- [ ] tests are passing
- [ ] user facing documentation has been updated
- [ ] you have configured git to use signed commits

## Publish Release

To make a release, do these steps with the new version:
1. Run `make ver=X.Y.Z version`
2. Update CHANGELOG.md
3. Merge changes
4. Run `make add-tag`
5. Run `make push-tag`
