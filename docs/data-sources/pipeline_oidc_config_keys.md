---
layout: "bitbucket"
page_title: "Bitbucket: bitbucket_pipeline_oidc_config_keys"
sidebar_current: "docs-bitbucket-data-pipeline-oidc-config-keys"
description: |-
  Provides a data for a Bitbucket pipeline OIDC Config Keys
---

# bitbucket\_user

Provdes a way to fetch data on a pipeline OIDC Config Keys.

## Example Usage

```hcl
data "bitbucket_pipeline_oidc_config_keys" "example" {
  workspace = "example"
}
```

## Argument Reference

The following arguments are supported:

* `workspace` - (Required) The workspace to fetch pipeline oidc config keys.

## Attributes Reference

* `oidc_config_keys` - The Json representing the OIDC config keys.
