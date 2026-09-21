resource "aws_db_subnet_group" "this" {
  name       = "${var.name}-db"
  subnet_ids = var.private_subnet_ids

  tags = { Name = "${var.name}-db" }
}

resource "aws_security_group" "db" {
  name        = "${var.name}-db"
  description = "PostgreSQL from the backend instances only"
  vpc_id      = var.vpc_id

  ingress {
    description     = "PostgreSQL from the backend security group"
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = [var.compute_security_group_id]
  }

  tags = { Name = "${var.name}-db" }
}

resource "aws_db_instance" "this" {
  identifier = "${var.name}-db"

  engine                     = "postgres"
  engine_version             = var.engine_version
  auto_minor_version_upgrade = true
  instance_class             = var.instance_class
  allocated_storage          = var.allocated_storage
  max_allocated_storage      = var.max_allocated_storage
  storage_type               = "gp3"
  storage_encrypted          = true

  db_name  = var.db_name
  username = var.username
  password = var.password

  db_subnet_group_name   = aws_db_subnet_group.this.name
  vpc_security_group_ids = [aws_security_group.db.id]
  publicly_accessible    = false
  multi_az               = var.multi_az

  # Restores are always to a NEW instance from a snapshot (docs/OPERATIONS.md), so the
  # live instance is never destroyed by a plan.
  deletion_protection       = true
  backup_retention_period   = 7
  backup_window             = "02:00-03:00"
  maintenance_window        = "sun:03:30-sun:04:30"
  copy_tags_to_snapshot     = true
  skip_final_snapshot       = false
  final_snapshot_identifier = "${var.name}-db-final"
  apply_immediately         = false

  tags = { Name = "${var.name}-db" }
}
