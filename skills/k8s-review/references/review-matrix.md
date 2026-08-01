# Kubernetes Review Matrix

Use the smallest applicable source or sanitized runtime evidence. Never interpret a missing object as healthy coverage.

| Dimension | Static evidence | Runtime evidence | Questions |
| --- | --- | --- | --- |
| Schema and APIs | Rendered objects, validator output, API versions | Server version and served-resource evidence | Are objects valid for the target and are APIs deprecated or near removal? |
| Security context | Pod and container security contexts, service accounts, RBAC references | Admission or controller errors from bounded events | Are privilege, identity, capability, filesystem, and namespace boundaries explicit and minimal? |
| Resources | Requests, limits, quotas, priority classes | Bounded node and container CPU or memory usage | Are requests schedulable and representative, and are limits tied to workload behavior? |
| Scheduling | Affinity, topology spread, tolerations, selectors | Node state, pending reasons, placement, recent eviction signals | Is placement intentional, resilient, and free of avoidable hot spots? |
| Autoscaling | HPA or VPA configuration and metric sources | Replica state, utilization, missing metrics, throttling indicators | Does scaling use available representative signals and safe bounds? |
| Probes and lifecycle | Startup, readiness, liveness, termination settings | Restarts, readiness, controller conditions, bounded logs | Do probes represent service health without causing avoidable disruption? |
| Disruption | Replicas, PodDisruptionBudgets, rollout strategy, topology | Availability and recent disruption events | Does planned disruption preserve the required availability? |
| Networking | Services, ingresses, NetworkPolicies, DNS assumptions | Endpoint state and bounded network-related events | Is exposure minimal and is intended connectivity represented? |
| Storage | PVCs, access modes, reclaim behavior, volume ownership | PV and PVC state, attachment or scheduling events | Are durability, topology, capacity, and deletion semantics understood? |
| Controllers | Owners, selectors, rollout strategy, Jobs and CronJobs | Conditions, desired versus ready state, failed Jobs | Are controllers converging and selectors or ownership unambiguous? |
| Upgrade risk | Cluster-version constraints, CRDs, webhook versions, deprecated APIs | Server version, warnings, controller compatibility evidence | What must change before the next supported upgrade? |
| Efficiency | Replica topology and declared resources | Requests versus usage, idle replicas, hot nodes or containers | Is there enough representative evidence for a reversible optimization experiment? |

For every row, record `verified`, `partial`, `not present`, or `not checked`, plus the exact evidence or gap. Keep snapshot age and workload seasonality visible when evaluating optimization.
