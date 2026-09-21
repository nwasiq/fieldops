# infra

Terraform for one environment, `envs/staging`, composed from four modules. It is **planned** from
this repository — offline, with no AWS account — and **never applied** from it.

## Modules

| Module | What it creates |
|---|---|
| `modules/network` | VPC, two public + two private subnets (one pair per availability zone), internet gateway, one NAT gateway, public and private route tables. Availability zones are a **variable**, never a data source. |
| `modules/compute` | Launch template from a variable `ami_id`, auto-scaling group (min/max/desired variables, rolling instance refresh), application load balancer + target group + HTTP listener, ALB and instance security groups, instance role limited to `ssm:GetParameter(s)` on `/fieldops/<env>/*`, pulling the one ECR repository, and Session Manager. User data reads `DATABASE_URL`, `JWT_SECRET` and `ENCRYPTION_KEY` from SSM at boot and runs the backend container on port 8080. |
| `modules/db` | RDS PostgreSQL 16, DB subnet group in the private subnets, a security group that admits the compute security group only, `deletion_protection = true`, 7-day automated backups, encrypted storage. The master password is a `sensitive` variable — never in code. |
| `modules/edge` | Private S3 bucket for the built frontend, CloudFront distribution with origin access control, `index.html` as the default root object, SPA routing (404 and 403 from S3 → `/index.html` with a 200). |

`envs/staging` wires them together: `variables.tf` is the full input surface, `staging.tfvars.example`
holds placeholder values (no real ids, no real secret), `providers.tf` configures the AWS provider
so that no credentials are needed to plan, and `outputs.tf` exposes the ALB host, the ASG name, the
DB address, the bucket and the distribution.

## Planning offline

Nothing here needs an AWS account or network access beyond the Terraform registry (for `init`):
the provider carries static placeholder keys and skips credential validation, account-id lookup,
the metadata endpoint and region validation; the state backend is `local`; no module contains a
`data "aws_*"` block (those need the API), and IAM documents are `jsonencode` literals.

```bash
cd infra/envs/staging
terraform init -input=false
terraform validate
terraform plan -input=false -var-file=staging.tfvars.example
terraform fmt -check -recursive ..      # from infra/ in CI
```

That is exactly what the `terraform` job in `.github/workflows/ci.yml` runs. Expected plan
summary: `Plan: 35 to add, 0 to change, 0 to destroy.` No `*.tfstate` is written: a plan against
an empty local state creates nothing. `.terraform.lock.hcl` is committed with hashes for
linux/darwin × amd64/arm64 so `init` is reproducible on a laptop and on the CI runner.

## Applying — not from here

`terraform apply` is deliberately not wired into any workflow, Makefile target or script in this
repository, and the committed provider configuration cannot apply: its keys are placeholders. Do
not add real credentials to `providers.tf` or to any `*.tfvars` — `infra/.gitignore` keeps
`*.tfvars` (except the `.example`) and `*_override.tf` out of the repository for that reason.

The real apply path is an operator workstation or a separate deploy pipeline holding the
credentials, working on a checkout of this configuration:

1. Drop a gitignored `infra/envs/staging/apply_override.tf` that overrides the `terraform` block
   with a remote backend (S3 bucket + DynamoDB lock table) and the `provider "aws"` block with a
   real credential source (an assumed role via OIDC in CI, or the operator's profile) and none of
   the `skip_*` flags. Terraform merges `*_override.tf` files last, so nothing in the committed
   files changes.
2. Copy `staging.tfvars.example` to the gitignored `staging.tfvars` with real values (region,
   an Amazon Linux 2023 AMI id, the ECR image and repository ARN, a globally unique bucket name).
   Supply the master password as `TF_VAR_db_password`, never in a file.
3. Create the three SSM SecureStrings before the first apply, since instances read them at boot:
   `/fieldops/staging/DATABASE_URL` (composed from the `db_address` and `db_port` outputs —
   `postgres://<user>:<password>@<address>:<port>/fieldops?sslmode=require`), `/fieldops/staging/JWT_SECRET`,
   `/fieldops/staging/ENCRYPTION_KEY`. With the default `aws/ssm` key no extra KMS grant is
   needed; with a customer-managed key set `ssm_kms_key_arn`.
4. `terraform init` (migrates state to the remote backend), `terraform plan -var-file=staging.tfvars`,
   review, `terraform apply` of that reviewed plan file.

A deploy after that is not a Terraform run: push the backend image to ECR, upload
`frontend/dist` to the bucket and invalidate the distribution, then start an instance refresh on
the ASG (`autoscaling_group_name` output). See `docs/OPERATIONS.md`.

## What the modules assume of the siblings

- **`backend/Dockerfile`** must exist: `docker-compose.yml` builds the `backend` service from it
  (profile `app`), and the ECR image the launch template runs is the same build. It listens on
  `PORT` (8080) and reads only the variables in `.env.example`.
- **`GET /api/health`** (configurable: `health_check_path`) must answer 200 without a token; the
  target group and the ASG's ELB health check probe it, and an instance that never answers is
  replaced.
- **`frontend`** must build with `npm ci && npm run build` into `dist/` and call the API at the
  relative path `/api/...`: the compose stack (`infra/docker/frontend.Dockerfile` + nginx on port
  3000, proxying `/api` to the backend service) and CloudFront both rely on that.
- The AMI is expected to be Amazon Linux 2023 (dnf, AWS CLI v2, systemd); user data installs
  docker if it is missing.

## Local full stack

```bash
cp .env.example .env
docker compose --profile app up --build     # db + backend :8080 + frontend :3000
```

The backend container gets `DATABASE_URL` pointing at the `db` service (overriding the localhost
value in `.env`) and waits for the database health check. The frontend image is built by
`infra/docker/frontend.Dockerfile` with `./frontend` as the build context, so `frontend/` itself
carries no Docker files.
