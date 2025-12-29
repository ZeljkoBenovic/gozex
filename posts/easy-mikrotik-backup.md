# Mikrotik Backups Made Easy
If you’ve ever managed a network of Mikrotik routers, you know the drill. 
Whether it’s a sudden hardware failure or a configuration tweak gone wrong, having a recent backup is the difference between a five-minute fix and a five-hour nightmare.

While RouterOS is powerful, managing backups across dozens or hundreds of devices can be a logistical headache. 
So, how do we cure this headache?  

Enter Gombak (GO Mikrotik BAcKup), a lightweight, open-source tool written in Go, designed specifically to automate and simplify the MikroTik backup process.

## Why Gombak?
Most network admins rely on custom scripts or manual exports. This is totally fine when there is only one, or a few routers to manage, 
but it is definitely not scalable to do *en masse*.   
Gombak makes this process very easy by providing a unified binary that handles discovery, execution, and retention. 
Since it’s built in Go, it’s fast, cross-platform, and has zero dependencies.

## Key Features at a Glance
* ***Three Operational Modes***: Flexible enough for one router or an entire ISP infrastructure.  
* ***L2TP Discovery***: Automatically finds routers connected via L2TP tunnels - perfect for "concentrator" setups.  
* ***Retention Management***: Automatically cleans up old backups so your storage doesn't bloat.   
* ***System Service Integration***: You can install it as a background service with a single command.  
* ***Config Flexibility***: Use CLI flags, Environment Variables, or YAML files.

## Getting Started
Gombak adapts to your network topology through three distinct modes of operation.   
There are 3 modes of operation:

1. **Single Mode**: Need a backup of a single Mikrotik device? You don’t even need a config file. Just run it straight from your terminal:
    ```bash
    gombak --single.host "192.168.88.1" --single.user "admin" --single.pass "password" --backup-dir "./backups"
    ```

2. **Multi Mode**: For a static list of routers, you can define a `config.yaml`. 
    This allows you to specify different credentials or SSH ports for every device in your inventory.
    Your config file would look like this: 
    ```yaml
    mode: multi
    backup-dir: "./backups"
    multi-router:
      - host: "10.0.0.1"
        username: "admin"
        password: "password"
      - host: "10.0.0.2"
        username: "admin"
        password: "password"
    ```

3. **The "Smart Discovery" (L2TP Mode)**: This is where Gombak truly shines for ISPs or large MSPs. 
   If you have a "concentrator" router where other routers connect via L2TP, Gombak can query the concentrator's API, discover the remote IP addresses of the tunnels, and back them all up automatically.   
    All you would need to do is to set up its configuration properly: 
    ```yaml
    mode: l2tp
    backup-dir: "./backups"
    discovery:
      hosts:
        - "10.0.0.1"
        - "10.0.0.2"
      username: "admin"
      password: "password"
    ```
## Automation Made Easy
Gombak comes with a very neat feature. Service installation. 
It includes built-in commands to install itself as a system service:  
* install: `gombak install -c /path/to/config.yaml`
* uninstall: `gombak uninstall`

Once installed, it will respect your defined backup frequency and backup retention, ensuring you always have a fresh backup without lifting a finger.

## Future 
Some features which will be added in the future: 
* **Prometheus metrics**: to enable observability of the backup process.
* **Cloud storage**: native support for saving files to cloud storage like S3 bucket.

## Conclusion
If you are looking for a "set it and forget it" solution for Mikrotik backups, Gombak is a fantastic choice. 
It moves away from the "do-it-your-self-script" of the past and provides a robust, professional CLI tool for modern network administration.

## Check it out on GitHub
Sounds useful? Check it out on [Gombak GitHub Repository](https://github.com/ZeljkoBenovic/gombak).