# Kubernetes Storage Operations Made Easy with kmon

**kmon** is a Kubernetes CLI and k9s plugin for PVC debugging, VolumeSnapshot restores, and scheduled storage snapshots — all in single commands.

Kubernetes does a lot of things well. Storage operations — debugging volume mount failures, restoring data from a VolumeSnapshot, scheduling recurring snapshots — are not among them.

The resources exist. `PersistentVolumeClaim`, `VolumeSnapshot`, `VolumeSnapshotContent` — the API surface is complete. But translating "I need to inspect what's actually on this PVC" or "restore this volume to yesterday's snapshot" into the correct sequence of `kubectl` commands, YAML manifests, and timing conditions is tedious enough that it tends not to happen until there's an incident.

**kmon** is a Kubernetes storage operations CLI and k9s plugin that automates these workflows into single commands.

## What kmon Does

kmon covers three storage operations that come up repeatedly in production Kubernetes environments:

1. **Debug pod from PVC** — attach a temporary pod to a live PVC to inspect its contents
2. **Restore from VolumeSnapshot** — create a new PVC from an existing snapshot with correct storage class binding
3. **Snapshot management** — take on-demand snapshots or deploy a CronJob for scheduled recurring snapshots

All three work as `kubectl` plugin invocations or directly through the k9s interface.

## Installing kmon

```bash
# Via go install
go install github.com/ZeljkoBenovic/kmon@latest

# Or download a release binary and place it in your $PATH
# Then verify kubectl picks it up
kubectl kmon --help
```

Once the binary is in `$PATH` with the `kubectl-kmon` naming convention, `kubectl kmon` works as a plugin command. For k9s, copy the binary to `~/.config/k9s/plugins/` and add the plugin configuration.

## Debugging a PVC Without Guessing

The most common storage debugging scenario: a pod fails to start because a volume mount fails, or application logs suggest the data isn't what you expect. To see what's actually on a PVC, you traditionally need to write a debug pod manifest, apply it, exec into it, poke around, then clean up.

kmon collapses that to:

```bash
kubectl kmon debug --pvc my-app-data -n production
```

This spins up a lightweight debug pod with the PVC mounted at `/data`, drops you into a shell, and deletes the pod when you exit. The PVC remains untouched.

You can specify the image if you need specific debugging tools:

```bash
kubectl kmon debug --pvc my-app-data -n production --image busybox
```

This works against live PVCs even if the original pod is still running, since kmon mounts the volume in `ReadOnlyMany` mode when supported by the storage class, falling back to a safe multi-attach check otherwise.

## Restoring a Volume from a VolumeSnapshot

When you need to recover data — whether for an actual incident or a dry-run of your DR procedure — the restore flow looks like this:

```bash
kubectl kmon restore \
  --snapshot my-app-data-snapshot-20250115 \
  --name my-app-data-restored \
  --namespace production
```

kmon handles the full sequence: it reads the snapshot's storage class and capacity, creates a correctly-sized PVC with `dataSource` pointing at the snapshot, waits for the PVC to bind, and confirms the restore is ready. No need to look up the original PVC spec or figure out which `VolumeSnapshotClass` to reference.

## Taking On-Demand Snapshots

For manual snapshots before a risky operation:

```bash
kubectl kmon snapshot \
  --pvc my-app-data \
  --name pre-migration-snapshot \
  --namespace production \
  --snapshot-class csi-aws-vsc
```

The `--snapshot-class` flag is optional if your cluster has a default `VolumeSnapshotClass` configured.

## Scheduling Recurring Snapshots

For automated backup schedules, kmon can deploy a Kubernetes CronJob that takes snapshots on a cron expression and retains the last N snapshots:

```bash
kubectl kmon schedule \
  --pvc my-app-data \
  --namespace production \
  --schedule "0 2 * * *" \
  --retain 7 \
  --snapshot-class csi-aws-vsc
```

This creates a CronJob that runs at 02:00 daily, takes a snapshot named `my-app-data-YYYYMMDD-HHMMSS`, and deletes snapshots beyond the retention count. No external backup operator required.

## k9s Integration

If you use k9s as your primary cluster UI, kmon registers itself as a plugin. With a PVC selected in k9s:

- `Shift-D` — launch debug pod
- `Shift-S` — take an on-demand snapshot
- `Shift-R` — restore from most recent snapshot

The plugin bindings are configurable in `~/.config/k9s/plugins.yaml`.

## When You Need This

kmon is most useful in three scenarios:

**During incidents** — quickly inspect PVC contents without writing YAML under pressure, or restore from the last known-good snapshot in minutes rather than looking up the procedure.

**During migrations** — snapshot before you move, restore if something goes wrong.

**For operational hygiene** — set up scheduled snapshots on stateful workloads that don't have an application-level backup mechanism. Databases running in Kubernetes with a nightly VolumeSnapshot CronJob are better protected than databases running without one.

## Conclusion

Kubernetes storage primitives are powerful, but the workflows around them don't need to be manual. kmon puts the most common storage operations — debug, snapshot, restore — behind clean commands that are fast enough to use during an incident and simple enough to use proactively.

Browse the [kmon source on GitHub](https://github.com/ZeljkoBenovic/kmon) and simplify your Kubernetes storage operations.
