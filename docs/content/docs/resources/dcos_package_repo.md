---
title: "dcos_package_repo"
type: docs
weight: 4
---

# Resource: dcos_package_repo

Installs a catalog repository on DC/OS that can be used to resolve service packages.

## Example Usage
```hcl
resource "dcos_package_repo" {
    name        = "Universe"
    url         = "https://universe.mesosphere.com/repo"
    volatile    = false
    index       = -1
}
```

## Argument Reference
The following arguments are supported

{{< tf_arguments >}}
    {{< tf_arg name="name" default="Universe" desc="the name of the repository." />}}
    {{< tf_arg name="url" default="https://universe.mesosphere.com/repo" desc="the url of the repository." />}}
    {{< tf_arg name="volatile" default="false" >}}
        if set to `true`, this repository will be removed when the resource is destroyed. If you are expecting the cluster to use this repository outside of your deployment, keep it `false`.
    {{</ tf_arg >}}
    {{< tf_arg name="index" default="-1" desc="defines the index where this repository will be installed at. This affects the resolution order for it's packages. If `-1`, it will be appended at the end (lowest priority), and if `0` it will be inserted at the beginning (highest priority)." />}}
{{</ tf_arguments >}}
