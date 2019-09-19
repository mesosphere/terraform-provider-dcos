package dcos

import (
	"context"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceDcosJob() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDcosJobRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Unique identifier for the job.",
			},
			"user": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The user to use to run the tasks on the agent.",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "A description of this job.",
			},
			"labels": {
				Type:        schema.TypeMap,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Attaching metadata to jobs can be useful to expose additional information to other services.",
			},
			"cmd": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The command that is executed. This value is wrapped by Mesos via `/bin/sh -c ${job.cmd}`. Either `cmd` or `args` must be supplied. It is invalid to supply both `cmd` and `args` in the same job.",
			},
			"args": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "An array of strings that represents an alternative mode of specifying the command to run. This was motivated by safe usage of containerizer features like a custom Docker ENTRYPOINT. Either `cmd` or `args` must be supplied. It is invalid to supply both `cmd` and `args` in the same job.",
			},
			"artifacts": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"uri": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "URI to be fetched by Mesos fetcher module.",
						},
						"executable": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Set fetched artifact as executable.",
						},
						"extract": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Extract fetched artifact if supported by Mesos fetcher module.",
						},
						"cache": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Cache fetched artifact if supported by Mesos fetcher module.",
						},
					},
				},
			},
			"docker": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"image": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The docker repository image name.",
						},
					},
				},
			},
			"env": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Environment variables (non secret)",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"secrets": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Any secrets that are necessary for the job",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"placement_constraint": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attribute": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The attribute name for this constraint.",
						},
						"operator": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The operator for this constraint.",
						},
						"value": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The value for this constraint.",
						},
					},
				},
			},
			"restart": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Defines the behavior if a task fails.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"active_deadline_seconds": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "If the job fails, how long should we try to restart the job. If no value is set, this means forever.",
						},
						"policy": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The policy to use if a job fails. NEVER will never try to relaunch a job. ON_FAILURE will try to start a job in case of failure.",
						},
					},
				},
			},
			"volume": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"container_path": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The path of the volume in the container.",
						},
						"host_path": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The path of the volume on the host.",
						},
						"mode": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Possible values are RO for ReadOnly and RW for Read/Write.",
						},
					},
				},
			},
			"cpus": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "The number of CPU shares this job needs per instance. This number does not have to be integer, but can be a fraction.",
			},
			"mem": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The amount of memory in MB that is needed for the job per instance.",
			},
			"disk": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "How much disk space is needed for this job. This number does not have to be an integer, but can be a fraction.",
			},
			"max_launch_delay": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of seconds until the job needs to be running. If the deadline is reached without successfully running the job, the job is aborted.",
			},
		},
	}
}

func dataSourceDcosJobRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ProviderState).Client
	ctx := context.TODO()

	jobId := d.Get("name").(string)

	metronome_v1_job, _, err := getDCOSJobInfo(jobId, client, ctx)
	if err != nil {
		return err
	}

	log.Printf("[TRACE] MetronomeV1Job: %+v", metronome_v1_job)

	d.SetId(metronome_v1_job.Id)
	d.Set("description", metronome_v1_job.Description)
	d.Set("cpus", metronome_v1_job.Run.Cpus)
	d.Set("mem", metronome_v1_job.Run.Mem)
	d.Set("disk", metronome_v1_job.Run.Disk)
	d.Set("max_launch_delay", metronome_v1_job.Run.MaxLaunchDelay)
	d.Set("args", metronome_v1_job.Run.Args)
	d.Set("cmd", metronome_v1_job.Run.Cmd)
	d.Set("user", metronome_v1_job.Run.User)

	return nil
}
