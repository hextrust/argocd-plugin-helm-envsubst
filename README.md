# argocd-helm-envsubst-plugin

## Usage
### Build dependency
```bash
$ argocd-helm-envsubst-plugin build 
Usage:
  argocd-helm-envsubst-plugin build [flags]

Flags:
      --helm-registry-secret-config-path string   Repository config, default to /helm-working-dir/plugin-repositories/repositories.yaml
  -h, --help                                      help for build
      --path string                               Path to the application
      --repository-path string                    Repository config, default to /helm-working-dir/
```

### Render helm template
```bash
Similar to helm template .

Usage:
  argocd-helm-envsubst-plugin render [flags]

Flags:
  -h, --help                  help for render
      --log-location string   Default to /tmp/argocd-helm-envsubst-plugin/
      --path string           Path to the application
```

## Development
```bash
# To rebuild, run and go into shell script
$ make build-run-sh

# Normal run
/ $ argocd-helm-envsubst-plugin

# With debug mode
/ $ DEBUG=1 argocd-helm-envsubst-plugin
```