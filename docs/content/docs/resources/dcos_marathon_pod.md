---
title: "dcos_marathon_pod"
variant: enterprise
type: docs
weight: 4
---

# Resource: dcos_marathon_pod
Provides a resource ...

## Example Usage

```hcl
# an example
provider "dcos" {
  cluster = "my-cluster"
}

```

## Argument Reference
The following arguments are supported

{{< tf_arguments >}}
    
    {{< tf_arg name="container"  desc="DC/OS secrets" />}}
    
    {{< tf_arg name="volume_mounts"  desc="DC/OS secrets" />}}
    
    {{< tf_arg name="name" required="true" desc="" />}}
    
    {{< tf_arg name="mount_path" required="true" desc="" />}}
    
    {{< tf_arg name="read_only"  desc="" />}}
    
    {{< tf_arg name="artifact"  desc="DC/OS secrets" />}}
    
    {{< tf_arg name="cache"  desc="" />}}
    
    {{< tf_arg name="executable"  desc="" />}}
    
    {{< tf_arg name="extract"  desc="" />}}
    
    {{< tf_arg name="uri"  desc="" />}}
    
    {{< tf_arg name="dest_path"  desc="" />}}
    
    {{< tf_arg name="name" required="true" desc="" />}}
    
    {{< tf_arg name="endpoints"  desc="DC/OS secrets" />}}
    
    {{< tf_arg name="name"  desc="" />}}
    
    {{< tf_arg name="container_port"  desc="" />}}
    
    {{< tf_arg name="host_port"  desc="" />}}
    
    {{< tf_arg name="protocol"  desc="" />}}
    
    {{< tf_arg name="labels"  desc="" />}}
    
    {{< tf_arg name="env"  desc="" />}}
    
    {{< tf_arg name="user"  desc="" />}}
    
    {{< tf_arg name="secret"  desc="DC/OS secrets" />}}
    
    {{< tf_arg name="env_var" required="true" desc="" />}}
    
    {{< tf_arg name="source" required="true" desc="" />}}
    
    {{< tf_arg name="lifecycle"  desc="DC/OS secrets" />}}
    
    {{< tf_arg name="kill_grace_period_seconds" required="true" desc="" />}}
    
    {{< tf_arg name="health_check"  desc="DC/OS secrets" />}}
    
    {{< tf_arg name="delay_seconds"  desc="" />}}
    
    {{< tf_arg name="http"  desc="DC/OS secrets" />}}
    
    {{< tf_arg name="path"  desc="" />}}
    
    {{< tf_arg name="scheme"  desc="" />}}
    
    {{< tf_arg name="endpoint"  desc="" />}}
    
    {{< tf_arg name="grace_period_seconds"  desc="" />}}
    
    {{< tf_arg name="interval_seconds"  desc="" />}}
    
    {{< tf_arg name="max_consecutive_failures"  desc="" />}}
    
    {{< tf_arg name="timeout_seconds"  desc="" />}}
    
    {{< tf_arg name="resources"  desc="DC/OS secrets" />}}
    
    {{< tf_arg name="mem"  desc="" />}}
    
    {{< tf_arg name="disk"  desc="" />}}
    
    {{< tf_arg name="gpus"  desc="" />}}
    
    {{< tf_arg name="cpus"  desc="" />}}
    
    {{< tf_arg name="exec"  desc="DC/OS secrets" />}}
    
    {{< tf_arg name="command_shell"  desc="" />}}
    
    {{< tf_arg name="image"  desc="DC/OS secrets" />}}
    
    {{< tf_arg name="kind"  desc="" />}}
    
    {{< tf_arg name="id"  desc="" />}}
    
    {{< tf_arg name="force_pull"  desc="" />}}
    
    {{< tf_arg name="labels"  desc="" />}}
    
    {{< tf_arg name="labels"  desc="" />}}
    
    {{< tf_arg name="scheduling"  desc="DC/OS secrets" />}}
    
    {{< tf_arg name="unreachable_strategy"  desc="DC/OS secrets" />}}
    
    {{< tf_arg name="inactive_after_seconds"  desc="" />}}
    
    {{< tf_arg name="expunge_after_seconds"  desc="" />}}
    
    {{< tf_arg name="backoff"  desc="DC/OS secrets" />}}
    
    {{< tf_arg name="backoff"  desc="" />}}
    
    {{< tf_arg name="backoff_factor"  desc="" />}}
    
    {{< tf_arg name="max_launch_delay"  desc="" />}}
    
    {{< tf_arg name="upgrade"  desc="DC/OS secrets" />}}
    
    {{< tf_arg name="minimum_health_capacity"  desc="" />}}
    
    {{< tf_arg name="maximum_over_capacity"  desc="" />}}
    
    {{< tf_arg name="kill_selection"  desc="" />}}
    
    {{< tf_arg name="volume"  desc="DC/OS secrets" />}}
    
    {{< tf_arg name="persistent"  desc="DC/OS secrets" />}}
    
    {{< tf_arg name="max_size"  desc="" />}}
    
    {{< tf_arg name="constraints"  desc="" />}}
    
    {{< tf_arg name="parameter"  desc="" />}}
    
    {{< tf_arg name="attribute" required="true" desc="" />}}
    
    {{< tf_arg name="operation" required="true" desc="" />}}
    
    {{< tf_arg name="type"  desc="" />}}
    
    {{< tf_arg name="size"  desc="" />}}
    
    {{< tf_arg name="name" required="true" desc="" />}}
    
    {{< tf_arg name="host" required="true" desc="" />}}
    
    {{< tf_arg name="secrets"  desc="DC/OS secrets" />}}
    
    {{< tf_arg name="secret_name" required="true" desc="" />}}
    
    {{< tf_arg name="env_var"  desc="" />}}
    
    {{< tf_arg name="source"  desc="" />}}
    
    {{< tf_arg name="user"  desc="" />}}
    
    {{< tf_arg name="marathon_service_url"  desc="By default we use the default DC/OS marathon serivce: service/marathon. But to support marathon on marathon the service url can be schanged." />}}
    
    {{< tf_arg name="executor_resources"  desc="DC/OS secrets" />}}
    
    {{< tf_arg name="cpus" required="true" desc="" />}}
    
    {{< tf_arg name="mem" required="true" desc="" />}}
    
    {{< tf_arg name="disk"  desc="" />}}
    
    {{< tf_arg name="name" required="true" desc="" />}}
    
    {{< tf_arg name="network"  desc="DC/OS secrets" />}}
    
    {{< tf_arg name="name"  desc="" />}}
    
    {{< tf_arg name="mode"  desc="" />}}
    
    {{< tf_arg name="labels"  desc="" />}}
    
    {{< tf_arg name="scaling"  desc="DC/OS secrets" />}}
    
    {{< tf_arg name="kind"  desc="" />}}
    
    {{< tf_arg name="instances"  desc="" />}}
    
    {{< tf_arg name="max_instances"  desc="" />}}
    
{{</ tf_arguments >}}

## Attributes Reference
 addition to all arguments above, the following attributes are exported:

 {{< tf_arguments >}}
     {{< tf_arg name="gid" desc="User ID to apply the grant on." />}}
 {{</ tf_arguments >}}
