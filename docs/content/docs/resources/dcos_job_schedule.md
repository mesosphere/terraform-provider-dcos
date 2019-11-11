---
title: "dcos_job_schedule"
variant: enterprise
type: docs
weight: 4
---

# Resource: dcos_job_schedule
provides a resource adding schedules to Metronome jobs.

## Example Usage

```hcl
# an example
provider "dcos" {
  cluster = "my-cluster"
}

data "dcos_job" "job" {
  name = "somejob"
}

resource "dcos_job_schedule" "jobsched" {
  dcos_job_id = "${data.dcos_job.job.name}"
  name        = "someschedule"
  cron        = "0,30 * * * *"
}

```

## Argument Reference
The following arguments are supported

{{< tf_arguments >}}

    {{< tf_arg name="name" required="true" desc="Unique identifier for the schedule." />}}

    {{< tf_arg name="cron" required="true" desc="Cron based schedule string" />}}

    {{< tf_arg name="concurrency_policy"  desc="Defines the behavior if a job is started, before the current job has finished. ALLOW will launch a new job, even if there is an existing run." />}}

    {{< tf_arg name="enabled"  desc="Defines if the schedule is enabled or not." />}}

    {{< tf_arg name="starting_deadline_seconds"  desc="The number of seconds until the job is still considered valid to start." />}}

    {{< tf_arg name="timezone"  desc="IANA based time zone string. See http://www.iana.org/time-zones for a list of available time zones." />}}

    {{< tf_arg name="dcos_job_id" required="true" desc="Unique identifier for the job." />}}

{{</ tf_arguments >}}
