---
title: "dcos_package_version"
type: docs
weight: 4
---

# Data Resource: dcos_package_version

Selects a package version from the given catalog repository.

## Example Usage

```hcl
resource "dcos_package_repo" "universe" { }

# Select a package version out of a repository
data "dcos_package_version" "jenkins-latest" {
    repo_url = "${dcos_package_repo.universe.url}"
    name     = "jenkins"
    version  = "latest"
    index    = -1
}
```

## Argument Reference

The following arguments are supported

{{< tf_arguments >}}
    {{< tf_arg name="repo_url" required="true" >}}
        the repository URL where to search for the package. This is typically the [`.url`]({{< relref "dcos_package_repo#url" >}}) output variable of a [`dcos_package_repo`]({{< relref "dcos_package_repo" >}}) resource
    {{</ tf_arg >}}
    {{< tf_arg name="name" required="ture" desc="the name of the package to resolve in the repository specified." />}}
    {{< tf_arg name="version" required="ture" desc="the version of the package to resolve (can be `latest` to resolve the latest available version)." />}}
    {{< tf_arg name="spec" output="true" >}}
        The package version specification that can be passed down to the [`.version_spec`]({{< relref "dcos_package_config#version_spec" >}}) argument of a [`dcos_package_config`]({{< relref "dcos_package_config" >}}) data resource.
    {{</ tf_arg >}}
{{</ tf_arguments >}}
