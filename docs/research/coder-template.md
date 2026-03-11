# Coder Development Environment

> Coder instance: https://code.dev.cauda.dev
> Provisioner: Kubernetes
> Docs: https://coder.com/docs

## Overview

Coder provides cloud development environments (CDEs) via Terraform templates. Each template defines the infrastructure (pods, containers) and the dev tools pre-installed.

## Template Structure

```
deploy/coder/
├── main.tf              # Terraform config (K8s pod, volumes, agents)
├── variables.tf         # Template variables (Go version, Node version, etc.)
├── outputs.tf           # Agent URLs
└── build/
    └── Dockerfile       # Custom dev image (optional)
```

## Kubernetes Template Basics

```hcl
terraform {
  required_providers {
    coder = { source = "coder/coder" }
    kubernetes = { source = "hashicorp/kubernetes" }
  }
}

data "coder_workspace" "me" {}
data "coder_workspace_owner" "me" {}

resource "coder_agent" "main" {
  os   = "linux"
  arch = "amd64"
  dir  = "/home/coder/lurkarr"

  startup_script = <<-EOT
    # Clone repo if not present
    if [ ! -d ~/lurkarr ]; then
      git clone https://github.com/lusoris/Lurkarr.git ~/lurkarr
    fi
    cd ~/lurkarr

    # Install Go tools
    go install go.uber.org/mock/mockgen@latest
    go install github.com/pressly/goose/v3/cmd/goose@latest

    # Install frontend deps
    cd frontend && npm ci && cd ..

    # Start PostgreSQL
    sudo pg_ctlcluster 17 main start
  EOT
}

resource "kubernetes_pod" "dev" {
  metadata {
    name      = "coder-${data.coder_workspace_owner.me.name}-${data.coder_workspace.me.name}"
    namespace = "coder-workspaces"
  }
  spec {
    container {
      name  = "dev"
      image = "ghcr.io/lusoris/lurkarr-devenv:latest"
      command = ["sh", "-c", coder_agent.main.init_script]
      env {
        name  = "CODER_AGENT_TOKEN"
        value = coder_agent.main.token
      }
      resources {
        requests = { cpu = "2", memory = "4Gi" }
        limits   = { cpu = "4", memory = "8Gi" }
      }
    }
  }
}

resource "coder_app" "vscode" {
  agent_id     = coder_agent.main.id
  slug         = "code-server"
  display_name = "VS Code"
  url          = "http://localhost:13337/?folder=/home/coder/lurkarr"
  icon         = "/icon/code.svg"
  subdomain    = true
}
```

## Dev Image Requirements

- Go 1.25+ toolchain
- Node.js 22 LTS
- PostgreSQL 17 client
- Docker CLI (for compose)
- Git, make, curl, jq
- mockgen, goose, golangci-lint

## VS Code Extensions to Pre-install

- golang.Go
- svelte.svelte-vscode
- bradlc.vscode-tailwindcss
- eamodio.gitlens
- ms-vscode.vscode-typescript-next
- github.vscode-pull-request-github

## Environment Variables

```
DATABASE_URL=postgres://lurkarr:lurkarr@localhost:5432/lurkarr?sslmode=disable
LISTEN_ADDR=:9705
LOG_LEVEL=debug
CSRF_KEY=dev-csrf-key-32-chars-minimum!!!
```

## Onboarding Steps

1. Create template in Coder UI at https://code.dev.cauda.dev
2. Upload main.tf (or push to linked GitHub repo)
3. Create workspace from template
4. Open VS Code Web or connect via SSH
5. Run `go test ./...` to verify setup
6. Run `docker compose up -d` for full stack

## GitHub Integration

- Coder already linked to GitHub
- Git credentials via Coder's built-in Git auth
- SSH keys auto-provisioned for the workspace
