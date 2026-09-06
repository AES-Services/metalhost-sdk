# Backups API

SDK v1.1.0 adds backup schedules, single-disk backups, and full-VM backup
manifests. These APIs require Metalhost backend v1.0.67 or later, including
migrations 00023–00024 and the lifecycle and Temporal workers.

## Single-disk backups

Use `StorageService.CreateSnapshot` with `source_disk`. Supply an
`Idempotency-Key` and reuse it when retrying an ambiguous failure.
`GetSnapshot` returns CREATING, READY or ERROR; `ListSnapshots` is
project-scoped and paginated. Snapshot metadata retains its original namespace
and network even after source deletion.

To restore, call `CreateDisk` with `from_snapshot` and a stable `disk_id`.
The new disk must use the snapshot's project, datacenter, network and tier;
its requested size must be at least the snapshot's reported size. Poll
`GetDisk` until AVAILABLE. Restoration never overwrites the source.

## Whole-VM backups

Use `SnapshotVirtualMachine`, then poll `GetVmSnapshot`. READY requires all
captured persistent disks, and `volumes` describes boot and data disks.
`consistency` distinguishes QUIESCED, CRASH_CONSISTENT and OFFLINE captures.
Disk-only snapshots are crash-consistent.

`CreateVirtualMachineFromBackup` returns an asynchronous operation. All captured
disks are restored before the new VM boots. The source VM, disks and existing
monthly commitment are unchanged. New resource billing and quotas apply.

## Policies

Use `CreateSnapshotSchedule` with a `schedule` containing:

```json
{
  "project_name": "projects/example",
  "target": "projects/example",
  "cadence": "DAILY",
  "hour_utc": 3,
  "minute_utc": 0,
  "retention_count": 3,
  "enabled": true
}
```

Target can instead be a VM resource name or disk resource name in that project.
HOURLY uses minute_utc; WEEKLY also uses weekday_utc (Sunday=0). Missing enabled
is false: set it explicitly. Reuse an Idempotency-Key for create retries.
The API stores fixed UTC times; the dashboard displays them in browser-local
time, which can shift when daylight saving time changes.

List policies by project; update using an explicit FieldMask. Pause via
`enabled=false`. Deleting a policy preserves existing backups. Keep-N applies
separately per source and policy after a new READY copy exists; it never removes
manual backups. Leave quota headroom for the next capture.

These are DC-local snapshots, not off-site disaster recovery or memory checkpoints.
