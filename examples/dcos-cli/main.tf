provider "dcos" {}

data "dcos_token" "current" {}
data "dcos_base_url" "current" {}

# This is an example of how to use the

resource "dcos_json_cli" "edgelb-pool" {
  # The `name` must uniquely identify this resource
  name  = "ping-lb-${count.index}"
  count = 10

  # The package whose cli to interact with
  package = "edgelb"

  # The configuration of this resource
  config = <<EOF
  {
    "name": "%NAME%",
    "apiVersion": "V2",
    "count": 2,
    "cpus": 1
  }
  EOF

  # The CRUD command-line arguments
  #
  # (Note that each command will be automatically prefixed
  #  with `dcos <package-cli> ...`)
  #
  # Before each command is invoked, the configuration JSON is serialized either
  # into a file or to a byte stream and fed into STDIN (or collected from STDOUT)
  # After the execution, the new config is read back
  #
  # The following macros can be used:
  #  %ID%       : Expands to the value of the `name` argument
  #  %CONFIG%   : Expands to a temporary filename with the configuration JSON
  # <%CONFIG%   : Writes the configuration JSON to STDIN
  # >%CONFIG%   : Reads the configuration JSON from STDOUT

  args_create = "create %CONFIG%"
  args_read   = "show --json %NAME% > %CONFIG%"
  args_delete = "delete %NAME%"
  args_update = "update %CONFIG%"

  # (Optional) Probe scripts
  #
  # The following scripts are executed in a shell and can be used to
  # check if a CRUD operation has completed.
  #
  # - Exiting with 0 indicates a successful operation
  # - Exiting with >0 indicates that the operation is not yet completed
  #
  # The following environment variables are defined:
  # - DCOS_CLI  : Expands to the full path to the DC/OS CLI
  # - DCOS_CMD  : Expands to `path/to/dcos-cli <package command>`

  probe_created = "$DCOS_CLI show %NAME%"
  probe_updated = ""
  probe_deleted = ""

  #
  # How long to wait for a probe to complete
  #

  probe_wait_duration = "5m"
}
