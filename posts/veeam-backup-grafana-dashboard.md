# Monitoring Veeam B&R with Govein

In the world of backup administration, "no news is good news" is a dangerous philosophy. 
You don't want to find out your backup jobs have been failing for three days only when you actually need to restore a critical server.

While Veeam Backup & Replication (VBR) provides email reports, modern IT teams need real-time, centralized observability and alerting. 
This is where **Govein** comes in - pumps your Veeam data directly into InfluxDB for stunning Grafana dashboards.

## What is Govein?

Govein is a Go based evolution of the popular community script most commonly used for Veeam monitoring
https://github.com/jorgedlcruz/veeam-backup-and-replication-grafana.  
It’s designed to be lightweight, container-friendly, and dead-simple to deploy. 
It scrapes the Veeam API and pushes stuff like job statuses, repositories usage percentage, etc. metrics into InfluxDB 2.0+.

## Why Go-Native?

If you've used older shell or PowerShell based exporters, you might have dealt with dependency or deployment issues.    
Govein offers:
* ***Container-First Design***: Easy to run in Docker or Kubernetes (Helm charts included!).
* ***Native Speed***: An executable, written in Go for minimal resource footprint on your monitoring host.
* ***Secrets Management***: Built-in support for environment variables, making it ideal for CI/CD and secure environments.
* ***Run on any host***: Various operating systems and CPU architectures are supported.

## Requirements

Before you start, make sure you have:

* `Veeam Backup & Replication 12+`: taking advantage of the latest Rest APIs.
* `InfluxDB 2.0+`: bucket created and ready.
* `Grafana`: to visualize the data.

## Fast-Track Configuration
Govein makes configuration intuitive. 
You can generate a starter template with `govein -export` or jump straight into a `config.yaml`:
```yaml
veeam:
    host: https://your-veeam-server:9419
    username: "veeam-admin"
    password: "secure-password"
    trust_self_signed_cert: true
influx:
    host: http://influxdb:8086
    token: "your-influx-token"
    org: "your-org"
    bucket: "veeam-metrics"
```

Once configured, simply run: `govein -config ./config.yaml`

## Make Veeam Dashboard Great Again
The real payoff happens in Grafana. 
Govein includes sample dashboards in the repository. 

Instead of digging through the Veeam console, your team can see:
* ***Job Success/Failure Ratios***: At a glance status of your entire environment.
* ***Data Transfer Trends***: Monitor how much data is being backed up daily.
* ***Storage Latency & Performance***: Identify bottlenecks before they impact your backup window.

## Deploying on Kubernetes
For those running on the all mighty orchestrator, we've got a neat Helm chart. 
Adding the Helm repo and deploying takes just two commands:
```bash 
helm repo add govein https://zeljkobenovic.github.io/govein
helm install my-govein govein/govein  
```

## Outro
Observability is the backbone of a reliable backup strategy. 
By moving your Veeam metrics into an InfluxDB/Grafana stack using Govein, you gain the visibility needed to move from reactive troubleshooting to proactive infrastructure management.

Ready to see your backups in a new light? Check out the [Govein GitHub Repo](https://github.com/ZeljkoBenovic/govein) and start scraping!