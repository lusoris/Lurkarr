terraform {
  required_providers {
    coder = {
      source = "coder/coder"
    }
    kubernetes = {
      source = "hashicorp/kubernetes"
    }
  }
}

# --- Variables ---

variable "namespace" {
  type        = string
  description = "Kubernetes namespace for workspaces"
  default     = "coder-workspaces"
}

variable "go_version" {
  type        = string
  description = "Go toolchain version"
  default     = "1.26.1"
}

variable "node_version" {
  type        = string
  description = "Node.js LTS version"
  default     = "22"
}

variable "cpu_request" {
  type        = string
  description = "CPU request for the workspace pod"
  default     = "2"
}

variable "memory_request" {
  type        = string
  description = "Memory request for the workspace pod"
  default     = "4Gi"
}

variable "cpu_limit" {
  type        = string
  description = "CPU limit for the workspace pod"
  default     = "4"
}

variable "memory_limit" {
  type        = string
  description = "Memory limit for the workspace pod"
  default     = "8Gi"
}

variable "home_disk_size" {
  type        = string
  description = "Persistent volume size for /home/coder"
  default     = "20Gi"
}

# --- Data Sources ---

data "coder_workspace" "me" {}
data "coder_workspace_owner" "me" {}

# --- Coder Agent ---

resource "coder_agent" "main" {
  os   = "linux"
  arch = "amd64"
  dir  = "/home/coder/lurkarr"

  env = {
    DATABASE_URL = "postgres://lurkarr:lurkarr@localhost:5432/lurkarr?sslmode=disable"
    LISTEN_ADDR  = ":9705"
    LOG_LEVEL    = "debug"
    CSRF_KEY     = "coder-dev-csrf-key-32-characters!"
  }

  startup_script_behavior = "blocking"

  startup_script = <<-EOT
    set -e

    # Clone repo if not present
    if [ ! -d ~/lurkarr ]; then
      git clone https://github.com/lusoris/Lurkarr.git ~/lurkarr
    fi
    cd ~/lurkarr

    # Install Go tools
    go install go.uber.org/mock/mockgen@latest
    go install github.com/pressly/goose/v3/cmd/goose@latest
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

    # Download Go module dependencies
    go mod download

    # Install frontend dependencies
    cd frontend && npm ci && cd ..

    # Wait for PostgreSQL sidecar
    until pg_isready -h localhost -U lurkarr; do sleep 1; done

    # Run database migrations
    goose -dir internal/database/migrations postgres "$DATABASE_URL" up
  EOT

  metadata {
    display_name = "CPU Usage"
    key          = "cpu"
    script       = "coder stat cpu"
    interval     = 10
    timeout      = 1
  }

  metadata {
    display_name = "Memory Usage"
    key          = "mem"
    script       = "coder stat mem"
    interval     = 10
    timeout      = 1
  }

  metadata {
    display_name = "Disk Usage"
    key          = "disk"
    script       = "coder stat disk --path /home/coder"
    interval     = 60
    timeout      = 1
  }
}

# --- Coder Apps ---

resource "coder_app" "vscode" {
  agent_id     = coder_agent.main.id
  slug         = "code-server"
  display_name = "VS Code"
  url          = "http://localhost:13337/?folder=/home/coder/lurkarr"
  icon         = "/icon/code.svg"
  subdomain    = true
}

resource "coder_app" "lurkarr" {
  agent_id     = coder_agent.main.id
  slug         = "lurkarr"
  display_name = "Lurkarr"
  url          = "http://localhost:9705"
  icon         = "/icon/widgets.svg"
  subdomain    = true
}

resource "coder_app" "grafana" {
  agent_id     = coder_agent.main.id
  slug         = "grafana"
  display_name = "Grafana"
  url          = "http://localhost:3000"
  icon         = "/icon/chart.svg"
  subdomain    = true
}

# --- Kubernetes Resources ---

resource "kubernetes_persistent_volume_claim" "home" {
  metadata {
    name      = "coder-${data.coder_workspace_owner.me.name}-${lower(data.coder_workspace.me.name)}-home"
    namespace = var.namespace
    labels = {
      "app.kubernetes.io/name"       = "coder-workspace"
      "app.kubernetes.io/instance"   = lower(data.coder_workspace.me.name)
      "app.kubernetes.io/managed-by" = "coder"
    }
  }
  wait_until_bound = false
  spec {
    access_modes = ["ReadWriteOnce"]
    resources {
      requests = {
        storage = var.home_disk_size
      }
    }
  }
}

resource "kubernetes_pod" "dev" {
  count = data.coder_workspace.me.start_count

  metadata {
    name      = "coder-${data.coder_workspace_owner.me.name}-${lower(data.coder_workspace.me.name)}"
    namespace = var.namespace
    labels = {
      "app.kubernetes.io/name"       = "coder-workspace"
      "app.kubernetes.io/instance"   = lower(data.coder_workspace.me.name)
      "app.kubernetes.io/managed-by" = "coder"
    }
  }

  spec {
    security_context {
      run_as_user = 1000
      fs_group    = 1000
    }

    # Main development container
    container {
      name  = "dev"
      image = "codercom/enterprise-base:ubuntu"

      command = ["sh", "-c", <<-EOT
        # Install Go
        curl -fsSL "https://go.dev/dl/go${var.go_version}.linux-amd64.tar.gz" | sudo tar -C /usr/local -xzf -
        echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' >> ~/.bashrc
        export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin

        # Install Node.js
        curl -fsSL https://deb.nodesource.com/setup_${var.node_version}.x | sudo -E bash -
        sudo apt-get install -y nodejs postgresql-client jq

        # Install code-server for VS Code web
        curl -fsSL https://code-server.dev/install.sh | sh -s -- --method=standalone
        code-server --install-extension golang.Go
        code-server --install-extension svelte.svelte-vscode
        code-server --install-extension bradlc.vscode-tailwindcss
        code-server --install-extension eamodio.gitlens
        code-server --install-extension ms-vscode.vscode-typescript-next
        code-server --install-extension github.vscode-pull-request-github
        code-server --auth none --port 13337 &

        # Start Coder agent
        ${coder_agent.main.init_script}
      EOT
      ]

      env {
        name  = "CODER_AGENT_TOKEN"
        value = coder_agent.main.token
      }

      resources {
        requests = {
          cpu    = var.cpu_request
          memory = var.memory_request
        }
        limits = {
          cpu    = var.cpu_limit
          memory = var.memory_limit
        }
      }

      volume_mount {
        name       = "home"
        mount_path = "/home/coder"
      }
    }

    # PostgreSQL sidecar
    container {
      name  = "postgres"
      image = "postgres:17-alpine"

      env {
        name  = "POSTGRES_USER"
        value = "lurkarr"
      }
      env {
        name  = "POSTGRES_PASSWORD"
        value = "lurkarr"
      }
      env {
        name  = "POSTGRES_DB"
        value = "lurkarr"
      }

      port {
        container_port = 5432
      }

      resources {
        requests = {
          cpu    = "250m"
          memory = "256Mi"
        }
        limits = {
          cpu    = "1"
          memory = "512Mi"
        }
      }

      volume_mount {
        name       = "pgdata"
        mount_path = "/var/lib/postgresql/data"
      }
    }

    volume {
      name = "home"
      persistent_volume_claim {
        claim_name = kubernetes_persistent_volume_claim.home.metadata[0].name
      }
    }

    volume {
      name = "pgdata"
      empty_dir {}
    }
  }
}
