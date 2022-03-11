---
layout: "bitbucket"
page_title: "Bitbucket: bitbucket_deploy_key"
sidebar_current: "docs-bitbucket-resource-deploy-key"
description: |-
  Provides a Bitbucket Deploy Key
---

# bitbucket\_deploy\_key

Provides a Bitbucket Deploy Key resource.

This allows you to manage your Deploy Keys for a repository.

## Example Usage

```hcl
resource "bitbucket_deploy_key" "test" {
  workspace  = "example"
  repository = "example"  
  key        = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIKqP3Cr632C2dNhhgKVcon4ldUSAeKiku2yP9O9/bDtY"
  label      = "test-key"
}
```

## Argument Reference

The following arguments are supported:

* `key` - (Required) The Deploy public key value in OpenDeploy format.
* `label` - (Optional) The user-defined label for the Deploy key

## Attributes Reference

* `key_id` - The Deploy key's ID.
* `comment` - The comment parsed from the Deploy key (if present)

## Import

Deploy Keys can be imported using their `workspace/repo-slug/key-id` ID, e.g.

```sh
terraform import bitbucket_deploy_key.key workspace/repo-slug/key-id
```
