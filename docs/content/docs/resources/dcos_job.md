---
title: "dcos_job"
variant: enterprise
type: docs
weight: 4
---

# Resource: dcos_job
Provides a resource creating metronome jobs on DC/OS.

Should be used together with `dcos_job_schedule`.

## Example Usage

```hcl
# an example
provider "dcos" {
  cluster = "my-cluster"
}

resource "dcos_job" "job" {
  name             = "testjob"
  cmd              = "echo foo"
  cpus             = 1
  mem              = 64
  disk             = 0
  user             = "root"
  description      = "the best description ever"
  max_launch_delay = 600

  docker {
    image = "ubuntu:latest"
  }

  placement_constraint {
    attribute = "host"
    operator  = "LIKE"
  }

  env {
    key    = "my_env_key"
    secret = "secret1"
  }

  env {
    key   = "cool_key"
    value = "cool_value"
  }

  secrets {
    secret1     = "/something"
    cool_secret = "something_else"
  }

  restart {
    active_deadline_seconds = 120
    policy                  = "NEVER"
  }

  artifacts {
    uri        = "http://downloads.mesosphere.com/robots.txt"
    extract    = false
    executable = true
    cache      = false
  }

  volume {
    container_path = "/mnt/test"
    host_path      = "/dev/null"
    mode           = "RW"
  }
}

resource "dcos_job_schedule" "jobsched" {
  dcos_job_id = "${dcos_job.job.name}"
  name        = "someschedule"
  cron        = "0,30 * * * *"
}

```

## Argument Reference
The following arguments are supported

{{< tf_arguments >}}

    {{< tf_arg name="ucr"  desc="" />}}

    {{< tf_arg name="image" required="true" desc="The ucr repository image name." />}}

    {{< tf_arg name="secrets"  desc="Any secrets that are necessary for the job" />}}

    {{< tf_arg name="placement_constraint"  desc="" />}}

    {{< tf_arg name="attribute" required="true" desc="The attribute name for this constraint." />}}

    {{< tf_arg name="operator" required="true" desc="The operator for this constraint." />}}

    {{< tf_arg name="value"  desc="The value for this constraint." />}}

    {{< tf_arg name="cpus" required="true" desc="The number of CPU shares this job needs per instance. This number does not have to be integer, but can be a fraction." />}}

    {{< tf_arg name="name" required="true" desc="Unique identifier for the job." />}}

    {{< tf_arg name="user"  desc="The user to use to run the tasks on the agent." />}}

    {{< tf_arg name="description"  desc="A description of this job." />}}

    {{< tf_arg name="labels"  desc="Attaching metadata to jobs can be useful to expose additional information to other services." />}}

    {{< tf_arg name="args"  desc="An array of strings that represents an alternative mode of specifying the command to run. This was motivated by safe usage of containerizer features like a custom Docker ENTRYPOINT. Either `cmd` or `args` must be supplied. It is invalid to supply both `cmd` and `args` in the same job." />}}

    {{< tf_arg name="env"  desc="Environment variables" />}}

    {{< tf_arg name="key" required="true" desc="The key/name of the variable" />}}

    {{< tf_arg name="value" required="true" desc="The value of the key/name" />}}

    {{< tf_arg name="secret"  desc="The name of the secret." />}}

    {{< tf_arg name="restart"  desc="Defines the behavior if a task fails." />}}

    {{< tf_arg name="active_deadline_seconds"  desc="If the job fails, how long should we try to restart the job. If no value is set, this means forever." />}}

    {{< tf_arg name="policy" required="true" desc="The policy to use if a job fails. NEVER will never try to relaunch a job. ON_FAILURE will try to start a job in case of failure." />}}

    {{< tf_arg name="mem" required="true" desc="The amount of memory in MB that is needed for the job per instance." />}}

    {{< tf_arg name="cmd"  desc="The command that is executed. This value is wrapped by Mesos via `/bin/sh -c ${job.cmd}`. Either `cmd` or `args` must be supplied. It is invalid to supply both `cmd` and `args` in the same job." />}}

    {{< tf_arg name="artifacts"  desc="" />}}

    {{< tf_arg name="cache"  desc="Cache fetched artifact if supported by Mesos fetcher module." />}}

    {{< tf_arg name="uri" required="true" desc="URI to be fetched by Mesos fetcher module." />}}

    {{< tf_arg name="executable"  desc="Set fetched artifact as executable." />}}

    {{< tf_arg name="extract"  desc="Extract fetched artifact if supported by Mesos fetcher module." />}}

    {{< tf_arg name="docker"  desc="" />}}

    {{< tf_arg name="image" required="true" desc="The docker repository image name." />}}

    {{< tf_arg name="volume"  desc="" />}}

    {{< tf_arg name="container_path" required="true" desc="The path of the volume in the container." />}}

    {{< tf_arg name="host_path" required="true" desc="The path of the volume on the host." />}}

    {{< tf_arg name="mode" required="true" desc="Possible values are RO for ReadOnly and RW for Read/Write." />}}

    {{< tf_arg name="secret"  desc="Allow for volume secrets if using UCR." />}}

    {{< tf_arg name="disk"  desc="How much disk space is needed for this job. This number does not have to be an integer, but can be a fraction." />}}

    {{< tf_arg name="max_launch_delay"  desc="The number of seconds until the job needs to be running. If the deadline is reached without successfully running the job, the job is aborted." />}}

{{</ tf_arguments >}}

## Attributes Reference
 addition to all arguments above, the following attributes are exported:

 {{< tf_arguments >}}
     {{< tf_arg name="gid" desc="User ID to apply the grant on." />}}
 {{</ tf_arguments >}}
