# cleta

__A cloud metadata server written in Go. Use as [CLI](#cli-usage).__

* Provision VMs using [cloud-init](https://cloudinit.readthedocs.io/en/latest/) images without a cloud provider.
* Runs on Linux and macOS.

## Supported Metadata Service APIs

* [DigitalOcean](https://developers.digitalocean.com/documentation/metadata/)
    * [Example Droplet](examples/sample-droplet.json)

### Planned
* [AWS](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html)
* [GCE](https://cloud.google.com/compute/docs/storing-retrieving-metadata)

## Storage Backends
* Filesystem Directories (JSON and YAML files).

### Planned
* Postgres
* MySQL / MariaDB
* MongoDB

## CLI usage

```bash
$ cleta serve \
    --metadata-bind-addr="169.254.169.254:80" \
    --metadata-store="dir" \
    --metadata-store-dir="a_path_to_vm_configs" \
    --metadata-store-dir="another_path_to_vm_configs" \
    --neighbor-table-refresh-interval=1ms
```