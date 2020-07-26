# Release

## Prerequisites

In order to perform releases you need to fulfill the following requirements.

1. Install GoReleaser. You can find instructions for your platform in the
   [official installation guide](https://goreleaser.com/install/).
2. Create and save a github token in a file at the path `~/.goreleaser/github-token`. If you are
   not keen on saving the token in your home folder, just note it somewhere
   else. You are going to need it when performing a release.
4. Install Docker. You can find instructions for your platform in the [official
   installation guide](https://docs.docker.com/install/).

Once these steps are completed, you are ready to perform a release. You don't
need to repeat these steps anymore.

## Perform a release

In order to perform a release, create a new Git tag in the form `vX.Y.Z` and
push it to the remote repository.

```
git tag -a vX.Y.Z -m 'Release version X.Y.Z'
git push origin vX.Y.Z
```



Once the tag has been correctly created and pushed, just invoke GoReleaser from
the root of the project.

```
goreleaser --rm-dist
```

If you decided not to save the GitHub Enterprise token in the file suggested in
the Prerequisites section, you need to explicitly pass the token to GoReleaser
via an environment variable.

```
GITHUB_TOKEN='your-github-token' goreleaser
```

## Snapshot release

```
goreleaser --snapshot --skip-publish --rm-dist
```
then you can list the docker images `docker images` and eventually push ems to the registry `docker push <image>`