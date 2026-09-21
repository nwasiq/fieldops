locals {
  ecr_registry      = split("/", var.backend_image)[0]
  ssm_parameter_arn = "arn:aws:ssm:${var.region}:*:parameter/fieldops/${var.env}/*"

  instance_policy_statements = concat(
    [
      {
        Sid      = "ReadFieldopsSecrets"
        Effect   = "Allow"
        Action   = ["ssm:GetParameter", "ssm:GetParameters"]
        Resource = local.ssm_parameter_arn
      },
      {
        Sid      = "EcrAuth"
        Effect   = "Allow"
        Action   = ["ecr:GetAuthorizationToken"]
        Resource = "*"
      },
      {
        Sid    = "EcrPullBackendImage"
        Effect = "Allow"
        Action = [
          "ecr:BatchGetImage",
          "ecr:GetDownloadUrlForLayer",
          "ecr:BatchCheckLayerAvailability",
        ]
        Resource = var.ecr_repository_arn
      },
    ],
    var.ssm_kms_key_arn == "" ? [] : [
      {
        Sid      = "DecryptFieldopsSecrets"
        Effect   = "Allow"
        Action   = ["kms:Decrypt"]
        Resource = var.ssm_kms_key_arn
      },
    ],
  )
}

# --- Security groups -------------------------------------------------------

resource "aws_security_group" "alb" {
  name        = "${var.name}-alb"
  description = "Public HTTP in, backend instances out"
  vpc_id      = var.vpc_id

  ingress {
    description = "HTTP from anywhere"
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = { Name = "${var.name}-alb" }
}

resource "aws_security_group" "instance" {
  name        = "${var.name}-backend"
  description = "Backend port from the ALB only; egress for ECR, SSM and RDS"
  vpc_id      = var.vpc_id

  ingress {
    description     = "Backend from the ALB"
    from_port       = 8080
    to_port         = 8080
    protocol        = "tcp"
    security_groups = [aws_security_group.alb.id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = { Name = "${var.name}-backend" }
}

# --- Instance role ---------------------------------------------------------

resource "aws_iam_role" "instance" {
  name = "${var.name}-backend-instance"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "ec2.amazonaws.com" }
      Action    = "sts:AssumeRole"
    }]
  })
}

resource "aws_iam_role_policy" "instance" {
  name = "${var.name}-backend-instance"
  role = aws_iam_role.instance.id

  policy = jsonencode({
    Version   = "2012-10-17"
    Statement = local.instance_policy_statements
  })
}

# Session Manager access for `docker logs fieldops-backend` (docs/OPERATIONS.md); no SSH.
resource "aws_iam_role_policy_attachment" "ssm_core" {
  role       = aws_iam_role.instance.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

resource "aws_iam_instance_profile" "instance" {
  name = "${var.name}-backend-instance"
  role = aws_iam_role.instance.name
}

# --- Launch template + ASG -------------------------------------------------

resource "aws_launch_template" "backend" {
  name_prefix   = "${var.name}-backend-"
  image_id      = var.ami_id
  instance_type = var.instance_type

  vpc_security_group_ids = [aws_security_group.instance.id]

  iam_instance_profile {
    arn = aws_iam_instance_profile.instance.arn
  }

  metadata_options {
    http_tokens                 = "required"
    http_put_response_hop_limit = 2
  }

  user_data = base64encode(templatefile("${path.module}/user_data.sh.tpl", {
    region        = var.region
    env           = var.env
    backend_image = var.backend_image
    ecr_registry  = local.ecr_registry
    seed_on_boot  = var.seed_on_boot ? "true" : "false"
  }))

  tag_specifications {
    resource_type = "instance"
    tags          = { Name = "${var.name}-backend" }
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_autoscaling_group" "backend" {
  name_prefix         = "${var.name}-backend-"
  min_size            = var.min_size
  max_size            = var.max_size
  desired_capacity    = var.desired_capacity
  vpc_zone_identifier = var.private_subnet_ids
  target_group_arns   = [aws_lb_target_group.backend.arn]

  health_check_type         = "ELB"
  health_check_grace_period = 300

  launch_template {
    id      = aws_launch_template.backend.id
    version = aws_launch_template.backend.latest_version
  }

  # A deploy is an instance refresh (docs/OPERATIONS.md): new launch-template version, rolling replace.
  instance_refresh {
    strategy = "Rolling"
    preferences {
      min_healthy_percentage = 50
    }
  }

  tag {
    key                 = "Name"
    value               = "${var.name}-backend"
    propagate_at_launch = true
  }

  lifecycle {
    create_before_destroy = true
  }
}

# --- Load balancer ---------------------------------------------------------

resource "aws_lb" "this" {
  name               = "${var.name}-alb"
  load_balancer_type = "application"
  internal           = false
  security_groups    = [aws_security_group.alb.id]
  subnets            = var.public_subnet_ids

  tags = { Name = "${var.name}-alb" }
}

resource "aws_lb_target_group" "backend" {
  name        = "${var.name}-backend"
  port        = 8080
  protocol    = "HTTP"
  vpc_id      = var.vpc_id
  target_type = "instance"

  deregistration_delay = 30

  health_check {
    path                = var.health_check_path
    matcher             = "200"
    interval            = 15
    timeout             = 5
    healthy_threshold   = 2
    unhealthy_threshold = 3
  }

  tags = { Name = "${var.name}-backend" }
}

resource "aws_lb_listener" "http" {
  load_balancer_arn = aws_lb.this.arn
  port              = 80
  protocol          = "HTTP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.backend.arn
  }
}
