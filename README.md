# `create-cli`

A command line client for downloading and configuring CREATE.

- [`create-cli`](#create-cli)
  - [Introduction](#introduction)
  - [Downloading and Installing](#downloading-and-installing)
  - [Usage](#usage)
    - [Download Create using `create-cli`](#download-create-using-create-cli)
    - [Configuring CREATE repositories using `create-cli`](#configuring-create-repositories-using-create-cli)
    - [Pushing to Remote Gitlab](#pushing-to-remote-gitlab)
  - [Contributing](#contributing)

## Introduction



## Downloading and Installing

To download the `create-cli`, get the desired binary version from the [Releases page](https://github.com/Capgemini/create-cli/releases) ensuring that you choose the correct binary type for your computer architecture.

## Usage

The following sections will detail how to use the `create-cli` to download and configure the CREATE source code.

### Download Create using `create-cli`

In order to download the CREATE repositories, you can either do it manually via git clone, or you can use the `create-cli` (recommended). Run the following to download the code and have it ready for the configuration step. You will need to give it a Github Personal Access Token so that it can correct clone (as it does not use SSH by default).

```shell
create-cli pre-install download \
    --pat $GITHUB_PERSONAL_ACCESS_TOKEN
```

For a better explanation of what the above variables are used for, run `create-cli pre-install download help`.

### Configuring CREATE repositories using `create-cli`

Once the CREATE repositories have been downloaded, you can configure them by using the `create-cli`. It is recommended to use the `create-cli` unless you know what you are doing. The `cli` is programmed to find specific variable placeholders and to replace them with real values that are auto-generated or provided by the user using the flags below. Run the following to configure the CREATE repository downloaded:

```shell
create-cli pre-install configure \
  --cloud-provider $CLOUD_PROVIDER \
  --create-url $CREATE_URL \
  --acme-reg-email $ACME_REG_EMAIL \
  --backstage-gitlab-token $BACKSTAGE_GITLAB_TOKEN \
  --gitlab-group $GITLAB_GROUP \
  --concourse-gitlab-token $CONCOURSE_GITLAB_TOKEN \
  --gitlab-pat-token $GITLAB_PAT_TOKEN
```

For a better explanation of what the above variables are used for, run `create-cli pre-install configure help`.

### Pushing to Remote Gitlab

Once the CREATE code has been configured, it can now be committed and pushed to the Gitlab Group/Subgroup of your choosing. The do this, run the following:

```shell
create-cli pre-install push \
    --cloud-provider $CLOUD_PROVIDER \
    --pat $GITLAB_PAT_TOKEN \
    --gitlab-group $GITLAB_GROUP
```

For a better explanation of what the above variables are used for, run `create-cli pre-install push help`.

## Contributing

To contribute to `create-cli` please find more information in the [contributing guides](./CONTRIBUTING.md).
