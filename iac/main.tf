# Infrastructure as Code — Blog Engine Sprint 1
# Provider: Generic (adapt to your cloud: Railway, Render, Fly.io, AWS, GCP)
# Last updated: 2026-05-30

terraform {
  required_version = ">= 1.7.0"
  required_providers {
    docker = {
      source  = "kreuzwerker/docker"
      version = "~> 3.0"
    }
  }
  # Remote state — configure your backend here
  # backend "s3" {
  #   bucket = "blog-engine-tfstate"
  #   key    = "sprint1/terraform.tfstate"
  #   region = "auto"
  # }
}

# ── Variables ─────────────────────────────────────────────────────────────────

variable "image_tag" {
  description = "Docker image tag to deploy (e.g. sprint1-abc1234)"
  type        = string
}

variable "db_url" {
  description = "PostgreSQL connection string"
  type        = string
  sensitive   = true
}

variable "jwt_secret" {
  description = "JWT signing secret (min 32 chars)"
  type        = string
  sensitive   = true
  default     = ""
}

variable "jwt_refresh_secret" {
  description = "JWT refresh signing secret"
  type        = string
  sensitive   = true
  default     = ""
}

variable "r2_account_id" {
  description = "Cloudflare R2 account ID"
  type        = string
  sensitive   = true
  default     = ""
}

variable "r2_access_key_id" {
  description = "Cloudflare R2 access key ID"
  type        = string
  sensitive   = true
  default     = ""
}

variable "r2_secret_access_key" {
  description = "Cloudflare R2 secret access key"
  type        = string
  sensitive   = true
  default     = ""
}

variable "r2_bucket_name" {
  description = "Cloudflare R2 bucket name"
  type        = string
  default     = "blog-engine-images"
}

variable "r2_public_url" {
  description = "Cloudflare R2 public URL for serving images"
  type        = string
  default     = ""
}

variable "smtp_host" {
  description = "SMTP host for transactional email"
  type        = string
  default     = ""
}

variable "smtp_port" {
  description = "SMTP port"
  type        = string
  default     = "587"
}

variable "smtp_user" {
  description = "SMTP username"
  type        = string
  sensitive   = true
  default     = ""
}

variable "smtp_pass" {
  description = "SMTP password"
  type        = string
  sensitive   = true
  default     = ""
}

variable "app_url" {
  description = "Frontend application URL"
  type        = string
  default     = "https://blog-engine.example.com"
}

variable "port" {
  description = "API server port"
  type        = string
  default     = "8080"
}

variable "registry" {
  description = "Container registry base URL"
  type        = string
  default     = "ghcr.io"
}

variable "image_repo" {
  description = "Image repository path"
  type        = string
  default     = "your-org/blog-engine/blog-engine-api"
}

# ── Frontend variables (added: frontend sprint) ───────────────────────────────

variable "frontend_image_tag" {
  description = "Frontend Docker image tag (e.g. frontend-abc1234)"
  type        = string
  default     = "latest"
}

variable "frontend_image_repo" {
  description = "Frontend image repository path"
  type        = string
  default     = "your-org/blog-engine/blog-engine-frontend"
}

variable "api_base_url" {
  description = "API base URL used by frontend (injected at build time)"
  type        = string
  default     = "https://api.blog-engine.example.com"
}

# ── Docker Provider ──────────────────────────────────────────────────────────

provider "docker" {}

# ── API Container ─────────────────────────────────────────────────────────────

resource "docker_image" "api" {
  name = "${var.registry}/${var.image_repo}:${var.image_tag}"
  pull_triggers = [var.image_tag]
}

resource "docker_container" "api" {
  name  = "blog-engine-api"
  image = docker_image.api.image_id
  restart = "unless-stopped"

  ports {
    internal = 8080
    external = 8080
  }

  env = [
    "DATABASE_URL=${var.db_url}",
    "JWT_SECRET=${var.jwt_secret}",
    "JWT_REFRESH_SECRET=${var.jwt_refresh_secret}",
    "R2_ACCOUNT_ID=${var.r2_account_id}",
    "R2_ACCESS_KEY_ID=${var.r2_access_key_id}",
    "R2_SECRET_ACCESS_KEY=${var.r2_secret_access_key}",
    "R2_BUCKET_NAME=${var.r2_bucket_name}",
    "R2_PUBLIC_URL=${var.r2_public_url}",
    "SMTP_HOST=${var.smtp_host}",
    "SMTP_PORT=${var.smtp_port}",
    "SMTP_USER=${var.smtp_user}",
    "SMTP_PASS=${var.smtp_pass}",
    "APP_URL=${var.app_url}",
    "PORT=${var.port}",
  ]

  healthcheck {
    test         = ["CMD", "curl", "-f", "http://localhost:8080/health"]
    interval     = "30s"
    timeout      = "10s"
    retries      = 3
    start_period = "15s"
  }

  labels {
    label = "sprint"
    value = "1"
  }

  labels {
    label = "deploy_tag"
    value = var.image_tag
  }
}

# ── Frontend Container (added: frontend sprint) ───────────────────────────────

resource "docker_image" "frontend" {
  name          = "${var.registry}/${var.frontend_image_repo}:${var.frontend_image_tag}"
  pull_triggers = [var.frontend_image_tag]
}

resource "docker_container" "frontend" {
  name    = "blog-engine-frontend"
  image   = docker_image.frontend.image_id
  restart = "unless-stopped"

  ports {
    internal = 80
    external = 3000
  }

  # nginx reverse-proxies /api/* to the API container on the same Docker network
  networks_advanced {
    name = docker_network.blog_engine.name
  }

  labels {
    label = "sprint"
    value = "frontend"
  }

  labels {
    label = "deploy_tag"
    value = var.frontend_image_tag
  }
}

# ── Shared Docker network (ensures frontend can reach API by container name) ──

resource "docker_network" "blog_engine" {
  name = "blog-engine-net"
}

# Attach API container to the shared network
resource "docker_container_network_attachment" "api_net" {
  container_id = docker_container.api.id
  network_id   = docker_network.blog_engine.id
}

# ── Outputs ───────────────────────────────────────────────────────────────────

output "deploy_tag" {
  value       = var.image_tag
  description = "Currently deployed image tag"
}

output "api_port" {
  value       = var.port
  description = "API server port"
}

output "r2_bucket" {
  value       = var.r2_bucket_name
  description = "Cloudflare R2 bucket for images"
}

output "frontend_deploy_tag" {
  value       = var.frontend_image_tag
  description = "Currently deployed frontend image tag"
}

output "frontend_port" {
  value       = "3000"
  description = "Frontend nginx port (external)"
}
