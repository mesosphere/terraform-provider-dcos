---
title: ""
type: docs
---

# DC/OS Provider

The DC/OS provider is used to interact with workload on DC/OS(dcos.io). Its configuration is the same as for [dcos-cli](github.com/dcos/dcos-cli)

## Example Usage

```hcl
provider "dcos" {
  dcos_url  = "<cluster url>"
  user      = "bootstrapuser"
  password  = "<secret dcos password>"
}

resource "dcos_marathon_pod" "simplepod" {
  name = "simplepod"

  scaling {
    kind      = "fixed"
    instances = 1
  }

  container {
    name = "sleep1"

    exec {
      command_shell = "sleep 1000"
    }

    resources {
      cpus = 0.1
      mem  = 32
    }
  }

  network {
    mode = "HOST"
  }
}

```

## Authentication and Configuration
The DC/OS provider is using the same config sources as the DC/OS CLI.

- Attached Cluster
- URL + Token
- URL + Username and Password
- Cluster Name

### Attached Cluster
The easiest solution is using the cluster you're attached to with `dcos cluster attach`

```hcl
provider "dcos" {}
```

The downside with this is that the user has to make sure being connected to the expected cluster.

### Cluster Name
If the user has locally setup his `dcos`-cli with `cluster setup <cluster url>` The name of a cluster can be used making sure terraform is using the expected cluster. The value is the same as for `dcos cluster attach`

```hcl
provider "dcos" {
  cluster = "my-dcos-production-cluster"
}
```

### Username and Password
__ENTERPRISE ONLY__

This method is using a username and password to authenticate against the DC/OS cluster. Be aware that this will not work with DC/OS Open Source.

```hcl
provider "dcos" {
  dcos_url  = "<cluster url>"
  user      = "bootstrapuser"
  password  = "<secret dcos password>"
}
```

### ACS Token
If you're using open source and don't want to use the attached cluster feature you have to specify the ACS token of a user (`dcos config show core.dcos_acs_token`). The token in combination with the cluster url give the dcos-provider access to your cluster.

```hcl
provider "dcos" {
  dcos_url       = "<cluster url>"
  dcos_acs_token = "<dcos_acs_token>"
}
```

## Argument Reference

{{< tf_arguments >}}
    {{< tf_arg name="dcos_acs_token"  desc="The DC/OS access token" />}}
    {{< tf_arg name="ssl_verify" default="true"  desc="Verify SSL connection. Can be set to false to ignore certificate errors" />}}
    {{< tf_arg name="dcos_url"  desc="The cluster URL. The same URL you reach the DC/OS UI" />}}
    {{< tf_arg name="cluster" >}}
        The cluster name configured in dcos-cli. You can obtain it using "dcos cluster list".
    {{</ tf_arg >}}
    {{< tf_arg name="user" ee="true" desc="The username to be used to connect to the DC/OS cluster" />}}
    {{< tf_arg name="password" ee="true" desc="The password to be used to connect to the DC/OS cluster" />}}
{{</ tf_arguments >}}
