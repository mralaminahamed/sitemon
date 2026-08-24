# Kubernetes manifests

Plain manifests for the full stack. Backing stores are ephemeral single
replicas — use managed services or StatefulSets with volumes in real clusters.

Apply:

    kubectl apply -f infra/k8s/sitemon.yaml

Set secrets before exposing:

    kubectl -n sitemon create secret generic sitemon-secrets \
      --from-literal=ANTHROPIC_API_KEY=... --from-literal=WEBHOOK_URL=... \
      --dry-run=client -o yaml | kubectl apply -f -

Images come from `.github/workflows/images.yml` (GHCR). Authored for reference;
validate against your cluster (`kubectl apply --dry-run=server`) before use.
