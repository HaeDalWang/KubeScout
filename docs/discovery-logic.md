# Discovery Logic: How KubeScout Finds Versions

KubeScout employs a **Hybrid Discovery Strategy** to ensure both accuracy for popular charts and flexibility for unknown ones. This approach minimizes "false positives" while maintaining the "Zero-Config" philosophy.

## 1. Prioritized Preset Registry (Accuracy First)

For widely used, critical infrastructure charts, KubeScout uses an internal **Preset Registry**. This maps common chart names directly to their specific repositories on Artifact Hub.

**Why?**
A generic search for "keda" might return a deprecated community fork or an unrelated chart with a similar name. Hardcoding the official repository ensures 100% accuracy.

**Current Presets:**
| Chart Name | Repository | Package |
|:---:|:---:|:---:|
| `argo-cd` | `argo` | `argo-cd` |
| `aws-load-balancer-controller` | `aws` | `aws-load-balancer-controller` |
| `karpenter` | `karpenter` | `karpenter` |
| `keda` | `kedacore` | `keda` |
| `cert-manager` | `cert-manager` | `cert-manager` |
| `ingress-nginx` | `ingress-nginx` | `ingress-nginx` |
| `prometheus` | `prometheus-community` | `prometheus` |
| `external-dns` | `external-dns` | `external-dns` |
| `n8n` | `n8n` | `n8n` |

## 2. Fallback Search Strategy (Zero-Config Flexibility)

If a chart is NOT in the preset list, KubeScout falls back to a smart search algorithm on Artifact Hub.

**Algorithm:**
1. **Search**: Queries Artifact Hub for the chart name (limit: 20 results).
2. **Filter**: Discards packages where the name does not match exactly.
3. **Rank**:
    - **Official Status**: Charts from "Official" repositories get top priority.
    - **Star Count**: Among non-official (or equally official) charts, the one with the most stars wins.
    - **Verified Publisher**: Used as a tie-breaker.

## 3. Future Extension

In strictly controlled environments, you may want to override these lookups. Future versions of KubeScout will support a `ConfigMap` to allow users to define their own `Chart -> Repo/Package` mapping, bypassing both presets and search.
