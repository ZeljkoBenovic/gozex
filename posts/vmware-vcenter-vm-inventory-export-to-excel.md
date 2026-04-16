# Exporting VMware vCenter VM Inventory to Excel

**vmex** is a Go CLI that queries VMware vCenter via the vSphere API and exports filtered VM inventory to formatted Excel workbooks — no PowerCLI required.

If you manage a VMware vSphere environment, you've almost certainly hit the same wall: you need a clean VM inventory for an audit, a capacity review, or a migration project, and vCenter's built-in export gives you a flat CSV with every column whether you want it or not.

That CSV rarely survives contact with a stakeholder. It needs reshaping, filtering, and formatting before it becomes a useful document. By the time you've done that manually twice, writing a tool to automate it starts looking like the sensible option.

**vmex** (VMware EXporter) queries the vSphere API and exports a filtered, formatted VM inventory directly to an Excel workbook — skipping the CSV intermediate entirely.

## What vmex Does

vmex connects to your vCenter instance using the govmomi vSphere SDK, queries the VM inventory with server-side filtering, and writes the result to an `.xlsx` file using structured column headers. You control which fields are included, which VMs are in scope, and where the file lands.

This replaces the manual export → open in Excel → delete columns → reformat → save cycle with a single command.

## Quick Start

```bash
# Install
go install github.com/ZeljkoBenovic/vmex@latest

# Basic export — all VMs in the inventory
vmex \
  --host vcenter.corp.example \
  --user administrator@vsphere.local \
  --password 'YourPassword' \
  --output inventory.xlsx
```

vmex will connect over HTTPS, enumerate all visible VMs, and write the workbook. The default column set covers the fields most useful for operational reporting:

- VM name, power state, and guest OS
- Assigned vCPUs and configured memory (GB)
- Provisioned and used disk (GB)
- Host, cluster, and datastore placement
- VMware Tools version and status
- IP addresses (all NICs)

## Filtering the Export

For large environments you rarely want everything. vmex supports filtering by datacenter, cluster, folder, and custom attribute:

```bash
# Export only VMs in a specific cluster
vmex \
  --host vcenter.corp.example \
  --user administrator@vsphere.local \
  --password 'YourPassword' \
  --cluster "Prod-Cluster-01" \
  --output prod-inventory.xlsx

# Export VMs matching a name prefix
vmex \
  --host vcenter.corp.example \
  --user administrator@vsphere.local \
  --password 'YourPassword' \
  --filter "web-*" \
  --output web-vms.xlsx
```

Filters are applied server-side via the vSphere API's property collector, which keeps the query efficient on inventories with thousands of VMs.

## Handling Self-Signed Certificates

Most internal vCenter deployments use self-signed TLS certificates. Pass `--insecure` to skip certificate verification:

```bash
vmex \
  --host vcenter.corp.example \
  --user administrator@vsphere.local \
  --password 'YourPassword' \
  --insecure \
  --output inventory.xlsx
```

For production use, the better option is to import the vCenter CA into your system trust store so `--insecure` isn't needed.

## Using Environment Variables

Credentials in command-line flags show up in shell history. vmex reads from environment variables too:

```bash
export VMEX_HOST=vcenter.corp.example
export VMEX_USER=administrator@vsphere.local
export VMEX_PASSWORD='YourPassword'

vmex --output inventory.xlsx
```

This makes it straightforward to wire into CI pipelines or cron jobs without embedding credentials in scripts.

## The Excel Output

The workbook uses a clean table format with frozen header rows and auto-sized columns. Numbers (CPU, RAM, disk) are formatted as numbers rather than strings, so Excel's sort and filter functions work correctly out of the box. Power state uses conditional cell colouring — green for powered on, grey for powered off — making it easy to scan a large inventory at a glance.

This is the detail that makes a difference when you're handing the file to someone who will be working with it in Excel rather than grep.

## Why Not Just Use PowerCLI?

PowerCLI is the standard tool for vSphere automation and it's excellent. But it requires a Windows machine (or a somewhat painful PowerShell Core setup on Linux), a VMware module install, and you still end up writing the export formatting logic yourself.

vmex is a single static binary with zero runtime dependencies. It runs on Linux, macOS, or Windows, needs no PowerShell, and produces formatted Excel output without any additional steps.

## Conclusion

For one-off audits, quarterly capacity reviews, or any reporting workflow that ends in an Excel file, vmex removes the friction between your vCenter inventory and a document someone will actually use.

Browse the [vmex source on GitHub](https://github.com/ZeljkoBenovic/vmex) and export your next vCenter inventory in seconds.
