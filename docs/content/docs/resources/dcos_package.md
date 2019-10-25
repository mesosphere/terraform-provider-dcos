---
title: "dcos_package"
type: docs
weight: 4
---

# Resource: dcos_package

Deploys (or upgrades) a service package on DC/OS.

## Example Usage
```hcl
data "dcos_package_config" "jenkins-config" {
    ...
}

# Deploy a package 
resource "dcos_package" "jenkins" {
    config          = "${data.dcos_package_config.jenkins-config.config}"
    app_id          = "/jenkins"
    wait            = true
    wait_duration   = "5m"
    sdk             = true
}
```

## Argument Reference
The following arguments are supported

{{< tf_arguments >}}
    {{< tf_arg name="config" required="true" >}}
        The configuration for the package to be deployed. This should be set to the [`.config`]({{< relref "dcos_package_config#config" >}}) output variable of a [`dcos_package_config`]({{< relref "dcos_package_config" >}}) data resource. (Note that the package name and version is specified in the package configuration).
    {{</ tf_arg >}}
    {{< tf_arg name="app_id" default="/<package-name>" >}}
        The name of the app to deploy on DC/OS
    {{</ tf_arg >}}
    {{< tf_arg name="wait" default="true" >}}
        When true, this resource will block any further action until the package is installed/uninstalled/updated. Set this to `false` to speed-up deployment, but only if you are not depending on any output variables of this resource.
    {{</ tf_arg >}}
    {{< tf_arg name="wait_duration" default="5m" >}}
        How long to wait for a blocking operation to complete. This can be any human-readable time expression (eg. “1h”, “10m”, “20s”)
    {{</ tf_arg >}}
    {{< tf_arg name="sdk" default="true" >}}
        When true, the provider will use the cosmos SDK API to update / restart the service. When false, any configuration change will cause the service to be uninstalled and re-installed.
    {{</ tf_arg >}}
{{</ tf_arguments >}}

## Updating Services

The dcos_package resource is smart enough to distinguish between configuration changes, version changes or name changes and can react accordingly.

### Service Redeployment

A service will be completely re-deployed (uninstalled and reinstalled) if any of the following changes have occurred:

* The package `name` has changed
* The `app_id` has changed
* The configuration has changed and `sdk=false` meaning that the service is not built using SDK, and therefore the provider cannot use the SDK api to update it.

### Service Reconfiguration

A service will be re-configured (“updated” using the cosmos API) if any of the following changes have occurred and sdk=true :

* The package `version` has changed
* The package configuration has changed
* The configuration `checksum` property has changed

### Service Restart

A service will be restarted (by force-restarting the “deploy” plan) if any of the following changes have occurred and `sdk=true`:

* Only the checksum property has changed (the rest of the configuration has remained intact)
