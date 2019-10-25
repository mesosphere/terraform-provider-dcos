---
title: "dcos_package_config"
type: docs
weight: 4
---

# Data Resource: dcos_package_config

Provides the configuration settings of a package before passing it to the [`dcos_package`]({{< relref "dcos_package" >}}) resource for the actual deployment.

## Example Usage

```hcl
resource "dcos_package_repo" "universe" { }

data "dcos_package_version" "jenkins-latest" {
    repo_url = "${dcos_package_repo.universe.url}"
    name     = "jenkins"
    version  = "latest"
    index    = -1
}

# Optional previous configuration to chain against
data "dcos_package_config" "previous" { ... }

# Configure a specific version of the package
data "dcos_package_config" "current" {
  version_spec      = "${data.dcos_package_version.jenkins-latest.spec}"
  extend            = "${data.dcos_package_config.previous.config}"
  autotype          = true
  checksum          = [ "version-1" ]

  section {
    path = "service"
    map = {
        cpus = 4
        mem = 2048
    }
  }
}
```

## Argument Reference

The following arguments are supported

{{< tf_arguments >}}
    {{< tf_arg name="version_spec" >}}
        The package name, version and repository meta-data. Assign here the value of the [`.spec`]({{< relref "dcos_package_version#spec" >}}) output variable of a [`dcos_package_version`]({{< relref "dcos_package_version" >}}) data resource.
    {{</ tf_arg >}}
    {{< tf_arg name="extend" >}}
        The previous configuration to chain. Assign here the value of the [`.config`](#config) output variable of another `dcos_package_config` resource.
    {{</ tf_arg >}}
    {{< tf_arg name="autotype" default="true" >}}
        If `true`, the provider will try to convert string JSON values from strings to their respective types (eg. “123” will become an integer 123)
    {{</ tf_arg >}}
    {{< tf_arg name="checksum" default="[]" >}}
        An array of arbitrary string expressions that can be used to calculate a unique checksum for this configuration.
    {{</ tf_arg >}}
    {{< tf_arg name="section" default="[]" >}}
        One or more configuration sections. Refer to [Configuration Sections](#configuration-sections) for more details.
    {{</ tf_arg >}}
    {{< tf_arg name="config" output="true" >}}
        This is an output (read-only) variable with the configuration meta-data of the package. Can be passed down to an [`extend`](#extend) property of another dcos_package_config resource, or to a dcos_package resource to deploy the service.
    {{</ tf_arg >}}
{{</ tf_arguments >}}

## Configuration Chaining

This data resource can be extended, enabling the user to share common parts. A configuration is extended by chaining the previous `.config` variable with the current `.extend` variable:

```hcl
data "dcos_package_config" "kafka-zookeeper" {
  ...
}

data "dcos_package_config" "kafka-zookeeper-kerberos" {
  extend = "${data.dcos_package_config.kafka-zookeeper.config}"
  ...
}
```

You are free to nest your configurations as desired, making sure that the final `dcos_package_config` data resource passed to the `dcos_package` resource contains a version information:

{{< mermaid class="text-center">}}
graph TD;
    subgraph Example 2;
    style cfg4 fill:#fff,stroke:#999;
    style cfg5 fill:#fff,stroke:#999;
    style cfg6 fill:#fff,stroke:#999;
    style cfg7 fill:#fff,stroke:#999;
    style cfg8 fill:#fff,stroke:#999;
    cfg4("Config<br />A");
    cfg4 -- .extend --> cfg5("Config<br />B");
    cfg4 -- .extend --> cfg6("Config<br />C");
    cfg5 -- .extend --> cfg7("Config<br />D");
    cfg6 -- .extend --> cfg8("Config<br />E");
    ver2>Version 1] -- .version_spec --> cfg7;
    ver3>Version 2] -- .version_spec --> cfg8;
    cfg7 -- .config --> depl3["Deployment<br/>A"];
    cfg8 -- .config --> depl4["Deployment<br/>B"];
    end;

    subgraph Example 1;
    style cfg1 fill:#fff,stroke:#999;
    style cfg2 fill:#fff,stroke:#999;
    style cfg3 fill:#fff,stroke:#999;
    ver1>Version] -- .version_spec --> cfg1("Config<br />A");
    cfg1 -- .extend  --> cfg2("Config<br/>B");
    cfg1 -- .extend --> cfg3("Config<br/>C");
    cfg2 -- .config --> depl1["Deployment<br/>A"];
    cfg3 -- .config --> depl2["Deployment<br/>B"];
    end;
{{< /mermaid >}}

You can attach a `version_spec` at any point of the chain. As seen in the example above, this is not required to happen on the root.

```hcl
data "dcos_package_version" "kafka-zookeeper" {
  ...
}

data "dcos_package_config" "kafka-zookeeper" {
  version_spec = "${data.dcos_package_version.kafka-zookeeper.spec}"
  ...
}
```

## Configuration Sections

In order to be able to provide configuration parameters in a format friendly to terraform we are using the concept of configuration “sections”. Each section is a value in a particular path in the configuration JSON object. There are three kinds of sections:

* [Object Sections](#object-section)
* [List Sections](#list-section)
* [Raw Sections](#raw-section)

Each section has a `path` property that specifies the JSON path where to inject it's properties. This property can point to an object at arbitrary depth. For example:

{{< columns >}}
The configuration:
```hcl
data "dcos_package_config" ... {
  section {
    path = "service"
    map {
      name = "foo"
    }
  }
}
```
<---> 
Will produce:
```json
{
  "service": {
    "name": "foo"
  }
}
```
{{< /columns >}}

{{< columns >}}
And the configuration:
```hcl
data "dcos_package_config" ... {
  section {
    path = "service.kdc.hosts"
    list = [
      "host1",
      "host2",
      "host3"
    ]
  }
}
```
<---> 
Will produce:
```json
{
  "service": {
    "kdc": {
      "hosts": [
        "host1",
        "host2",
        "host3"
      ]
    }
  }
}
```
{{< /columns >}}

### Object Section

The object section is equivalent to a JSON object. The section properties will be mapped 1:1 to the resulting JSON.

```hcl
  section {
    path = "service.config"
    map = {
        name = "foo"
        cpus = 3
        ...
    }
  }
```

{{< tf_arguments prefix="object-" >}}
    {{< tf_arg name="path" required="true" >}}
        The path in the resulting object where to inject the object properties.
    {{</ tf_arg >}}
    {{< tf_arg name="map" required="true" >}}
        An object with an arbitrary set of (key/value) pairs that will be inserted to the target path.
    {{</ tf_arg >}}
{{</ tf_arguments >}}


### List Section

The list section is equivalent to a JSON array, with the only exception that you must specify scalar values. 

```hcl
  section {
    path = "service.kdc.hosts"
    list = [
        "host-1",
        "host-2",
        "host-3"
    ]
  }
```

{{< tf_arguments prefix="list-" >}}
    {{< tf_arg name="path" required="true" >}}
        The path in the resulting object where to inject the object properties.
    {{</ tf_arg >}}
    {{< tf_arg name="list" required="true" >}}
        An list of scalar values (eg. strings, booleans, numbers) that will be inserted to the target path.
    {{</ tf_arg >}}
{{</ tf_arguments >}}

### Raw Section

If the map or list values are not fitting in your case, you can fall-back to raw JSON string.

```hcl
  section {
    path = "service"
    json = <<EOF
    {
      "placement_constraint": "[[\"@zone\",\"GROUP_BY\",\"3\"]]"
    }
    EOF
  }
```

{{< tf_arguments prefix="raw-" >}}
    {{< tf_arg name="path" required="true" >}}
        The path in the resulting object where to inject the object properties.
    {{</ tf_arg >}}
    {{< tf_arg name="raw" required="true" >}}
        An valid JSON string that will be inserted to the target path. This can be an object, an array or a scalar value.
    {{</ tf_arg >}}
{{</ tf_arguments >}}

