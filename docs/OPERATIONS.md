# Operations

## Deploy

`infra/` describes one environment (`infra/envs/staging`): a VPC, an auto-scaling group of backend instances behind an ALB pulling the backend image from ECR, an RDS PostgreSQL instance, an S3 bucket + CloudFront distribution serving the built frontend. Secrets (`DATABASE_URL`, `JWT_SECRET`, `ENCRYPTION_KEY`) live in SSM Parameter Store as SecureStrings and are read by the instance at boot, never baked into the image or user data.

`terraform plan` runs with no AWS credentials: the provider is configured with `skip_credentials_validation`, `skip_requesting_account_id` and `skip_metadata_api_check`, and the backend is local. `terraform apply` is deliberately not wired anywhere in this repository.

The CI in this repository (`.github/workflows/ci.yml`) is gates only — tests, lint scripts and the offline `terraform plan` — and holds no credentials. The deploy pipeline lives with whoever holds the AWS account, outside this repository, and follows `infra/README.md`: build and push the backend image to ECR, upload the frontend bundle to the S3 bucket and invalidate the CloudFront distribution, then start an ASG instance refresh. A deploy is "done" when the refresh reports every instance healthy, not when a workflow goes green.

## Logs

Backend logs are structured JSON on stdout (`log/slog`). On an instance: `docker logs fieldops-backend`. Every request line carries `request_id`, `path`, `status`, `duration_ms`, and `user_id` once authenticated. A 500 always logs the wrapped error with the path.

## Backups and restore

RDS takes automated snapshots daily with a 7-day retention. Restoring is a **new** instance from a snapshot, a smoke test against it, then a `DATABASE_URL` switch in SSM and an instance refresh — never an in-place restore.

Locally the same discipline applies: `make db-dump` writes `backups/fieldops-<timestamp>.sql`; `make db-restore FILE=…` restores into a fresh database and prints row counts per table so the restore can be verified before it is trusted.
