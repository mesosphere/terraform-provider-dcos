---
title: "dcos_service_http_request"
type: docs
weight: 4
---

# Resource: dcos_service_http_request

Performs an HTTP request to a DC/OS service, exposed through the Admin Router.

## Example Usage
```hcl
# POST some data to an HTTP service 
resource "dcos_service_http_request" "some_config" {
  service_name = "my-service"
  path         = "/configure"
  method       = "POST"

  header {
    name  = "Content-Type"
    value = "application/json"
  }

  body = <<EOF
    {
        "setting-a": "foo",
        "setting-b": "bar"
    }
EOF
}
```

## Argument Reference
The following arguments are supported

{{< tf_arguments >}}
    {{< tf_arg name="service_name" required="true" >}}
        The name of the DC/OS service to target. This is the same as the marathon app id without the leading path (eg. `my-group/my-service`).
    {{</ tf_arg >}}
    {{< tf_arg name="path" default="/" >}}
        The path under the exposed service endpoint to access (including the leading slash). The final URL will be composed as: `https://<cluster>/service/<service><path>`.
    {{</ tf_arg >}}
    {{< tf_arg name="method" default="GET" >}}
        The HTTP method to use for performing this request. Can be any valid HTTP verb (eg. "GET", "POST", "PUT", "DELETE") etc.
    {{</ tf_arg >}}
    {{< tf_arg name="header" >}}
        One or more headers to include in the request. The Authorisation header is always included.
    {{</ tf_arg >}}
    {{< tf_arg name="body" default="" >}}
        The raw body of the HTTP request.
    {{</ tf_arg >}}
    {{< tf_arg name="run_on" default="create" >}}
        Specify *when* to perform the HTTP request. There are two options:
        * `create`: Perform the HTTP request when the resource is created.
        * `delete`: Perform the HTTP request when the resource is deleted.
    {{</ tf_arg >}}
{{</ tf_arguments >}}
