# EC2 Fleet Command Execution Without Opening SSH

**aws-commander** is a Go CLI for fleet-wide EC2 command execution via AWS SSM Run Command — no SSH keys, no open ports, target instances by ID or tag.

The standard playbook for running commands across a fleet of EC2 instances used to be: open port 22, distribute SSH keys, write a loop over your inventory. It works, but it trades security for convenience. Every instance becomes a potential entry point, key rotation is painful, and the blast radius of a compromised key spans your entire fleet.

AWS Systems Manager (SSM) Run Command eliminates all of that. No open ports, no SSH keys, no bastion hosts. But the SSM console is clunky for bulk operations — and the AWS CLI interface for Run Command is verbose enough to make scripting it a chore.

**aws-commander** wraps SSM Run Command in a purpose-built CLI that makes fleet-wide execution feel like `ssh`, not like wrestling with JSON payloads.

## How aws-commander Uses AWS SSM Run Command

aws-commander talks to the AWS SSM API directly. As long as your EC2 instances have the `AmazonSSMManagedInstanceCore` IAM policy attached and the SSM Agent running (it's pre-installed on Amazon Linux and most modern AMIs), you're ready. No inbound security group rules needed — the agent polls outbound over HTTPS.

You target instances in two ways:

- **By instance ID** — for surgical operations on specific machines
- **By tag** — for fleet-wide operations on a logical group (e.g. all instances tagged `Environment=production`)

## Running Shell Commands, Scripts, and Ansible Playbooks

The simplest use case: run a shell command across every instance tagged as a web server.

```bash
aws-commander run \
  --tag "Role=web-server" \
  --command "systemctl status nginx"
```

For ad-hoc diagnostics that are too complex for a one-liner, point it at a script file:

```bash
aws-commander run \
  --tag "Environment=production" \
  --script ./scripts/collect-diagnostics.sh
```

And for infrastructure automation workflows, it supports Ansible playbooks as first-class input:

```bash
aws-commander run \
  --tag "Environment=staging" \
  --playbook ./playbooks/patch-and-restart.yml
```

Under the hood, aws-commander invokes the appropriate SSM Run Command document (`AWS-RunShellScript` for shell commands and scripts, `AWS-RunAnsiblePlaybook` for Ansible) and streams the output back to your terminal as it arrives.

## Targeting Specific Instances

When you need to target a known instance directly:

```bash
aws-commander run \
  --instance-id i-0abc123def456789 \
  --command "df -h"
```

Or multiple instance IDs at once:

```bash
aws-commander run \
  --instance-ids i-0abc123def456789,i-0def456789abc123 \
  --command "uptime"
```

## IAM Setup

aws-commander needs an IAM identity with `ssm:SendCommand` and `ssm:GetCommandInvocation` permissions. A minimal policy:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ssm:SendCommand",
        "ssm:GetCommandInvocation",
        "ssm:ListCommandInvocations"
      ],
      "Resource": "*"
    }
  ]
}
```

For production use, scope the `Resource` to specific instance ARNs or SSM documents rather than wildcards.

## Why Not Just Use the AWS CLI?

You absolutely can. Here's the equivalent `aws` CLI call for running a single command on a tagged fleet:

```bash
aws ssm send-command \
  --document-name "AWS-RunShellScript" \
  --targets "Key=tag:Role,Values=web-server" \
  --parameters "commands=[\"systemctl status nginx\"]" \
  --output json \
  --query "Command.CommandId"
```

Then you need to poll `aws ssm list-command-invocations` with that CommandId to get results. aws-commander handles all of this — targeting, document selection, polling, and output aggregation — in a single command invocation.

## Security Benefits

This approach is worth emphasising: aws-commander works entirely within the SSM trust boundary. Your instances never need port 22 open. The SSM Agent authenticates outbound to the AWS API using the instance's IAM role — no credentials on disk, no key pairs to manage, no lateral movement risk from a stolen SSH key.

Combined with IMDSv2 enforcement and strict IAM role scoping, this is the correct posture for any EC2 fleet that takes its security surface area seriously.

## Conclusion

If you're still SSHing into EC2 instances to run fleet-wide commands, aws-commander and SSM Run Command are the upgrade path. You get better security, less operational overhead, and a clean audit trail in CloudTrail — for free.

Browse the [aws-commander source on GitHub](https://github.com/ZeljkoBenovic/aws-commander) and start running SSM commands across your EC2 fleet.
