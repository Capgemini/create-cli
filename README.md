# create-cli

A command line client for CREATE.

## Releasing `create-cli`

In order to release the `create-cli`, simply create the correct release and tags via the release page. It is important that you also modify the `cmd/version.go` file to have the correct vesrion that you want to release. This isn't the best way of ensuring we set the correct versions for the `cli` but for now it will do.

Example:
From:

```golang
fmt.Println("create-cli 1.0.0.0")
```

To:

```golang
fmt.Println("create-cli 1.0.0.1")
```

Once this has been committed to main alongside the changes associated with the new version, publish the new release and tag and the Github Actions workflow will ensure that the container is built and published to `ghcr.io/cd-create/create-cli:[VERSION]`, alongside the Binaries being attached to the Release Assets.
