---
title: "dcos_marathon_app"
variant: enterprise
type: docs
weight: 4
---

# Resource: dcos_marathon_app
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
    
    {{< tf_arg name="args"  desc="" />}}
    
    {{< tf_arg name="cpus"  desc="" />}}
    
    {{< tf_arg name="fetch"  desc="" />}}
    
    {{< tf_arg name="extract"  desc="" />}}
    
    {{< tf_arg name="uri"  desc="" />}}
    
    {{< tf_arg name="cache"  desc="" />}}
    
    {{< tf_arg name="executable"  desc="" />}}
    
    {{< tf_arg name="user"  desc="" />}}
    
    {{< tf_arg name="executor"  desc="" />}}
    
    {{< tf_arg name="dcos_framework"  desc="" />}}
    
    {{< tf_arg name="plan_path"  desc="" />}}
    
    {{< tf_arg name="timeout"  desc="Timeout in seconds to wait for a framework to complete deployment" />}}
    
    {{< tf_arg name="is_framework"  desc="" />}}
    
    {{< tf_arg name="constraints"  desc="" />}}
    
    {{< tf_arg name="attribute"  desc="" />}}
    
    {{< tf_arg name="operation"  desc="" />}}
    
    {{< tf_arg name="parameter"  desc="" />}}
    
    {{< tf_arg name="disk"  desc="" />}}
    
    {{< tf_arg name="labels"  desc="" />}}
    
    {{< tf_arg name="networks"  desc="" />}}
    
    {{< tf_arg name="name"  desc="" />}}
    
    {{< tf_arg name="mode" required="true" desc="" />}}
    
    {{< tf_arg name="labels"  desc="" />}}
    
    {{< tf_arg name="version"  desc="" />}}
    
    {{< tf_arg name="accepted_resource_roles"  desc="" />}}
    
    {{< tf_arg name="container"  desc="" />}}
    
    {{< tf_arg name="volumes"  desc="" />}}
    
    {{< tf_arg name="host_path"  desc="" />}}
    
    {{< tf_arg name="secret"  desc="" />}}
    
    {{< tf_arg name="mode"  desc="" />}}
    
    {{< tf_arg name="external"  desc="" />}}
    
    {{< tf_arg name="options"  desc="" />}}
    
    {{< tf_arg name="name"  desc="" />}}
    
    {{< tf_arg name="provider"  desc="" />}}
    
    {{< tf_arg name="persistent"  desc="" />}}
    
    {{< tf_arg name="type"  desc="" />}}
    
    {{< tf_arg name="size"  desc="" />}}
    
    {{< tf_arg name="max_size"  desc="" />}}
    
    {{< tf_arg name="container_path"  desc="" />}}
    
    {{< tf_arg name="port_mappings"  desc="" />}}
    
    {{< tf_arg name="name"  desc="" />}}
    
    {{< tf_arg name="network_names"  desc="" />}}
    
    {{< tf_arg name="container_port"  desc="" />}}
    
    {{< tf_arg name="host_port"  desc="" />}}
    
    {{< tf_arg name="service_port"  desc="" />}}
    
    {{< tf_arg name="protocol"  desc="" />}}
    
    {{< tf_arg name="labels"  desc="" />}}
    
    {{< tf_arg name="type"  desc="" />}}
    
    {{< tf_arg name="docker"  desc="" />}}
    
    {{< tf_arg name="parameters"  desc="" />}}
    
    {{< tf_arg name="key"  desc="" />}}
    
    {{< tf_arg name="value"  desc="" />}}
    
    {{< tf_arg name="privileged"  desc="" />}}
    
    {{< tf_arg name="force_pull_image"  desc="" />}}
    
    {{< tf_arg name="image" required="true" desc="" />}}
    
    {{< tf_arg name="pull_config"  desc="" />}}
    
    {{< tf_arg name="secret"  desc="" />}}
    
    {{< tf_arg name="gpus"  desc="" />}}
    
    {{< tf_arg name="backoff_seconds"  desc="" />}}
    
    {{< tf_arg name="dependencies"  desc="" />}}
    
    {{< tf_arg name="env"  desc="" />}}
    
    {{< tf_arg name="secrets"  desc="" />}}
    
    {{< tf_arg name="unreachable_strategy"  desc="" />}}
    
    {{< tf_arg name="inactive_after_seconds"  desc="" />}}
    
    {{< tf_arg name="expunge_after_seconds"  desc="" />}}
    
    {{< tf_arg name="app_id" required="true" desc="" />}}
    
    {{< tf_arg name="mem"  desc="" />}}
    
    {{< tf_arg name="backoff_factor"  desc="" />}}
    
    {{< tf_arg name="health_checks"  desc="" />}}
    
    {{< tf_arg name="grace_period_seconds"  desc="" />}}
    
    {{< tf_arg name="interval_seconds"  desc="" />}}
    
    {{< tf_arg name="port_index"  desc="" />}}
    
    {{< tf_arg name="port"  desc="" />}}
    
    {{< tf_arg name="timeout_seconds"  desc="" />}}
    
    {{< tf_arg name="ignore_http_1xx"  desc="" />}}
    
    {{< tf_arg name="max_consecutive_failures"  desc="" />}}
    
    {{< tf_arg name="protocol"  desc="" />}}
    
    {{< tf_arg name="delay_seconds"  desc="" />}}
    
    {{< tf_arg name="path"  desc="" />}}
    
    {{< tf_arg name="command"  desc="" />}}
    
    {{< tf_arg name="value"  desc="" />}}
    
    {{< tf_arg name="port_definitions"  desc="" />}}
    
    {{< tf_arg name="port"  desc="" />}}
    
    {{< tf_arg name="name"  desc="" />}}
    
    {{< tf_arg name="labels"  desc="" />}}
    
    {{< tf_arg name="protocol"  desc="" />}}
    
    {{< tf_arg name="max_launch_delay_seconds"  desc="" />}}
    
    {{< tf_arg name="kill_selection"  desc="" />}}
    
    {{< tf_arg name="marathon_service_url"  desc="By default we use the default DC/OS marathon serivce: service/marathon. But to support marathon on marathon the service url can be schanged." />}}
    
    {{< tf_arg name="cmd"  desc="" />}}
    
    {{< tf_arg name="instances"  desc="" />}}
    
    {{< tf_arg name="require_ports"  desc="" />}}
    
    {{< tf_arg name="upgrade_strategy"  desc="" />}}
    
    {{< tf_arg name="minimum_health_capacity"  desc="" />}}
    
    {{< tf_arg name="maximum_over_capacity"  desc="" />}}
    
{{</ tf_arguments >}}

## Attributes Reference
 addition to all arguments above, the following attributes are exported:

 {{< tf_arguments >}}
     {{< tf_arg name="gid" desc="User ID to apply the grant on." />}}
 {{</ tf_arguments >}}
