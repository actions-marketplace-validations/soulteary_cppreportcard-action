# Marketplace publishing

This repository contains one root Action (`action.yml`), its Go CLI, supporting
scripts, documentation and tests. The repository must remain public. The listing
name is **C++ Report Card**, with an `award` icon on a blue background.

## Prepare a release

1. Merge the desired changes and confirm the default-branch CI passes.
2. Create and push a new immutable version tag, for example:

   ```sh
   git tag v1.0.1
   git push origin v1.0.1
   ```

3. Wait for the Release workflow. It runs the same full CI as pull requests,
   including the local Action and real C++ analyzers, then builds six CLI archives
   and checksums and saves a **draft**, rather than publishing it automatically.
   A rerun refreshes assets only while the release remains a draft; published
   releases are never overwritten by this workflow.

## Publish the Marketplace listing

Open the draft from [Releases](https://github.com/soulteary/cppreportcard-action/releases).
In the release form:

- Select **Publish this Action to the GitHub Marketplace**.
- Check that GitHub accepts the metadata and listing name. Names must be globally
  unique; a local check cannot reserve or guarantee availability.
- Choose **Code quality** as the primary category; **Code review** is an optional
  secondary category.
- Confirm the release tag and notes, then publish the release.

The repository owner must accept the GitHub Marketplace Developer Agreement if
GitHub requests it, and must have two-factor authentication enabled. Do not put
these account-level steps into CI. GitHub's supported publication flow exposes
Marketplace enrollment and categories in the release form; `gh release create`
alone does not set them.

[Official publishing instructions](https://docs.github.com/en/actions/how-tos/create-and-publish-actions/publish-in-github-marketplace)

## Promote the stable major tag

After publishing and verifying the exact version, move `v1` to it deliberately.
Do not move an immutable version tag such as `v1.0.0` or `v1.0.1`.

```sh
git fetch origin --tags
previous_v1=$(git rev-parse refs/tags/v1)
git tag -f v1 v1.0.1
git push origin refs/tags/v1 --force-with-lease="refs/tags/v1:$previous_v1"
```

Keep the previous major-tag SHA for rollback. Consumers needing immutable
versions should pin a full commit SHA or an exact release tag.

For the already-published `v1.0.0`, the owner can alternatively edit that release
and enable Marketplace listing without moving any tag. The v1.0.1 draft provides
the improved publishing workflow and documentation.

## Check the listing

Verify that the published Marketplace page links to this repository and shows
`uses: soulteary/cppreportcard-action@v1`. Follow the listing URL GitHub supplies;
do not assume a slug before the name has passed GitHub's uniqueness check.
