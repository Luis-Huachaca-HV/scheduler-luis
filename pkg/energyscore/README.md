# Energy Score Plugin

## Overview

The Energy Score plugin is a scheduling plugin that optimizes pod placement based on node energy consumption metrics. This plugin helps reduce the overall energy footprint of your Kubernetes cluster by preferring nodes with better energy efficiency characteristics.

## Maturity Level

- [x] 💡 Sample (for demonstrating and inspiring purpose)
- [ ] 👶 Alpha (used in companies for pilot projects)
- [ ] 👦 Beta (used in companies and developed actively)
- [ ] 👨 Stable (used in companies for production workloads)

## How it Works

The EnergyScore plugin implements the `Score` extension point of the Kubernetes scheduler framework. During the scoring phase, it evaluates each candidate node and assigns a score based on energy-related metrics. Nodes with better energy efficiency receive higher scores, making them more likely to be selected for pod placement.

### Scoring Algorithm

The plugin considers the following factors when scoring nodes:

1. **Node Energy Efficiency**: Evaluates the energy consumption characteristics of the node
2. **Current Load**: Considers the current resource utilization of the node
3. **Power States**: Takes into account whether the node supports and uses energy-saving features

The final score is calculated using a weighted formula that can be tuned using the plugin configuration.

## Configuration

The plugin can be configured through the `KubeSchedulerConfiguration` file:

```yaml
apiVersion: kubescheduler.config.k8s.io/v1
kind: KubeSchedulerConfiguration
profiles:
  - schedulerName: energy-scheduler
    plugins:
      score:
        enabled:
          - name: EnergyScore
            weight: 2
    pluginConfig:
      - name: EnergyScore
        args:
          weightMultiplier: 1.0
```

### Configuration Parameters

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `weight` | int | Weight of the plugin in the scoring phase (1-100) | 1 |
| `weightMultiplier` | float | Additional multiplier to adjust scores | 1.0 |

## Usage

To use the Energy Score plugin, deploy a pod with the appropriate scheduler name:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: energy-aware-app
spec:
  schedulerName: energy-scheduler
  containers:
    - name: app
      image: nginx:latest
      resources:
        requests:
          cpu: "100m"
          memory: "128Mi"
```

## Installation

See the main [README](../../README.md#energy-scheduler---guía-de-instalación-y-uso) for detailed installation instructions.

## Future Improvements

- Integration with real-time power monitoring systems
- Support for dynamic power capping
- Carbon-aware scheduling based on grid energy sources
- Historical power consumption analysis
- Multi-objective optimization with performance and energy trade-offs

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## License

This plugin is licensed under the Apache License 2.0. See the [LICENSE](../../LICENSE) file for details.
