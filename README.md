[![Go Report Card](https://goreportcard.com/badge/kubernetes-sigs/scheduler-plugins)](https://goreportcard.com/report/kubernetes-sigs/scheduler-plugins) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/kubernetes-sigs/scheduler-plugins/blob/master/LICENSE)

# Scheduler Plugins

Repository for out-of-tree scheduler plugins based on the [scheduler framework](https://kubernetes.io/docs/concepts/scheduling-eviction/scheduling-framework/).

This repo provides scheduler plugins that are exercised in large companies.
These plugins can be vendored as Golang SDK libraries or used out-of-box via the pre-built images or Helm charts.
Additionally, this repo incorporates best practices and utilities to compose a high-quality scheduler plugin.

## Install

Container images are available in the official scheduler-plugins k8s container registry. There are two images one
for the kube-scheduler and one for the controller. See the [Compatibility Matrix section](#compatibility-matrix)
for the complete list of images.

```shell
docker pull registry.k8s.io/scheduler-plugins/kube-scheduler:$TAG
docker pull registry.k8s.io/scheduler-plugins/controller:$TAG
```

You can find [how to install release image](doc/install.md) here.

## Plugins

The kube-scheduler binary includes the below list of plugins. They can be configured by creating one or more
[scheduler profiles](https://kubernetes.io/docs/reference/scheduling/config/#multiple-profiles).

* [Capacity Scheduling](pkg/capacityscheduling/README.md)
* [Coscheduling](pkg/coscheduling/README.md)
* [Node Resources](pkg/noderesources/README.md)
* [Node Resource Topology](pkg/noderesourcetopology/README.md)
* [Preemption Toleration](pkg/preemptiontoleration/README.md)
* [Trimaran (Load-Aware Scheduling)](pkg/trimaran/README.md)
* [Network-Aware Scheduling](pkg/networkaware/README.md)
* [Energy Score](pkg/energyscore/README.md)

Additionally, the kube-scheduler binary includes the below list of sample plugins. These plugins are not intended for use in production
environments.

* [Cross Node Preemption](pkg/crossnodepreemption/README.md)
* [Pod State](pkg/podstate/README.md)
* [Quality of Service](pkg/qos/README.md)

## Energy Scheduler - Guía de Instalación y Uso

El plugin **EnergyScore** es un scheduler personalizado que optimiza la programación de pods basándose en métricas de consumo energético de los nodos.

### Prerrequisitos

- Kubernetes cluster v1.28+ 
- `kubectl` configurado para acceder al cluster
- Go 1.22+ (para compilar desde fuente)

### Opción 1: Como Segundo Scheduler (Recomendado para Pruebas)

Esta opción mantiene el scheduler default de Kubernetes y añade el energy-scheduler como un scheduler adicional.

#### 1. Compilar la Imagen del Scheduler

```bash
# Clonar el repositorio
git clone https://github.com/Luis-Huachaca-HV/scheduler-luis.git
cd scheduler-luis

# Compilar las imágenes
make local-image

# O usar el script de build
./hack/build-images.sh
```

#### 2. Cargar la Imagen en el Cluster

```bash
# Para Kind
kind load docker-image localhost:5000/scheduler-plugins/kube-scheduler:latest

# Para Minikube
minikube image load localhost:5000/scheduler-plugins/kube-scheduler:latest
```

#### 3. Crear el ConfigMap con la Configuración

```bash
kubectl create configmap energy-scheduler-config \
  --from-file=energy-score-config.yaml \
  -n kube-system
```

#### 4. Desplegar el Energy Scheduler

```bash
kubectl apply -f manifests/energy-scheduler/rbac.yaml
kubectl apply -f manifests/energy-scheduler/deployment.yaml
```

#### 5. Verificar que el Scheduler está Corriendo

```bash
kubectl get pods -n kube-system | grep energy-scheduler
kubectl logs -n kube-system -l component=energy-scheduler
```

### Opción 2: Reemplazando el Default Scheduler

> ⚠️ **Advertencia**: Esta opción reemplaza el scheduler por defecto. Úsala con precaución.

#### 1. Acceder al Nodo Control-Plane

```bash
# Para Kind
sudo docker exec -it $(sudo docker ps | grep control-plane | awk '{print $1}') bash

# Para clusters con kubeadm
ssh usuario@control-plane-node
```

#### 2. Hacer Backup del Scheduler Original

```bash
cp /etc/kubernetes/manifests/kube-scheduler.yaml /etc/kubernetes/kube-scheduler.yaml.backup
```

#### 3. Crear la Configuración del Energy Scheduler

```bash
cat > /etc/kubernetes/energy-score-config.yaml << 'EOF'
apiVersion: kubescheduler.config.k8s.io/v1
kind: KubeSchedulerConfiguration
clientConnection:
  kubeconfig: "/etc/kubernetes/scheduler.conf"
leaderElection:
  leaderElect: true
profiles:
  - schedulerName: default-scheduler
    plugins:
      score:
        enabled:
          - name: EnergyScore
            weight: 2
    pluginConfig:
      - name: EnergyScore
        args:
          weightMultiplier: 1.0
EOF
```

#### 4. Modificar el Manifiesto del kube-scheduler

Editar `/etc/kubernetes/manifests/kube-scheduler.yaml` y actualizar la imagen y los argumentos:

```yaml
spec:
  containers:
    - name: kube-scheduler
      image: registry.k8s.io/scheduler-plugins/kube-scheduler:v0.31.8
      # O tu imagen local: localhost:5000/scheduler-plugins/kube-scheduler:latest
      command:
        - /bin/kube-scheduler
        - --authentication-kubeconfig=/etc/kubernetes/scheduler.conf
        - --authorization-kubeconfig=/etc/kubernetes/scheduler.conf
        - --config=/etc/kubernetes/energy-score-config.yaml
        - --v=5
      volumeMounts:
        - mountPath: /etc/kubernetes/scheduler.conf
          name: kubeconfig
          readOnly: true
        - mountPath: /etc/kubernetes/energy-score-config.yaml
          name: energy-config
          readOnly: true
  volumes:
    - hostPath:
        path: /etc/kubernetes/scheduler.conf
        type: FileOrCreate
      name: kubeconfig
    - hostPath:
        path: /etc/kubernetes/energy-score-config.yaml
        type: FileOrCreate
      name: energy-config
```

#### 5. Verificar el Scheduler

```bash
kubectl get pods -n kube-system | grep kube-scheduler
kubectl logs -n kube-system kube-scheduler-<node-name>
```

### Uso del Energy Scheduler

#### Programar Pods con el Energy Scheduler

Para que un pod use el Energy Scheduler, especifica `schedulerName` en el spec:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: energy-aware-pod
spec:
  schedulerName: energy-scheduler  # Usar si instalaste como segundo scheduler
  containers:
    - name: app
      image: nginx:latest
      resources:
        requests:
          cpu: "100m"
          memory: "128Mi"
```

Puedes probar con el archivo de ejemplo:

```bash
kubectl apply -f test-pod.yaml
kubectl get pods -o wide
```

### Configuración del Plugin EnergyScore

El archivo [energy-score-config.yaml](energy-score-config.yaml) permite personalizar el comportamiento:

| Parámetro | Descripción | Default |
|-----------|-------------|---------||
| `weight` | Peso del plugin en la fase de scoring (1-100) | 1 |
| `weightMultiplier` | Multiplicador adicional para ajustar scores | 1.0 |

### Desinstalar el Energy Scheduler

#### Si Instalaste como Segundo Scheduler

```bash
kubectl delete -f manifests/energy-scheduler/deployment.yaml
kubectl delete -f manifests/energy-scheduler/rbac.yaml
kubectl delete configmap energy-scheduler-config -n kube-system
```

#### Si Reemplazaste el Default Scheduler

1. Acceder al nodo control-plane
2. Restaurar el backup:

```bash
mv /etc/kubernetes/kube-scheduler.yaml.backup /etc/kubernetes/manifests/kube-scheduler.yaml
rm /etc/kubernetes/energy-score-config.yaml
```

3. Verificar que el scheduler default está corriendo:

```bash
kubectl get pods -n kube-system | grep kube-scheduler
```

### Troubleshooting

**El scheduler no inicia:**

```bash
# Ver logs del scheduler
kubectl logs -n kube-system -l component=energy-scheduler

# Verificar eventos
kubectl get events -n kube-system --field-selector reason=FailedScheduling
```

**Los pods quedan en Pending:**

```bash
# Verificar el schedulerName del pod
kubectl get pod <pod-name> -o jsonpath='{.spec.schedulerName}'

# Ver eventos del pod
kubectl describe pod <pod-name>
```

**Verificar que el plugin está habilitado:**

```bash
kubectl logs -n kube-system energy-scheduler-xxx | grep -i "energy"
```

## Compatibility Matrix

The below compatibility matrix shows the k8s client package (client-go, apimachinery, etc) versions
that the scheduler-plugins are compiled with.

The minor version of the scheduler-plugins matches the minor version of the k8s client packages that
it is compiled with. For example scheduler-plugins `v0.18.x` releases are built with k8s `v1.18.x`
dependencies.

The scheduler-plugins patch versions come in two different varieties (single digit or three digits).
The single digit patch versions (e.g., `v0.18.9`) exactly align with the k8s client package
versions that the scheduler plugins are built with. The three digit patch versions, which are built
on demand, (e.g., `v0.18.800`) are used to indicated that the k8s client package versions have not
changed since the previous release, and that only scheduler plugins code (features or bug fixes) was
changed.

| Scheduler Plugins | Compiled With k8s Version | Container Image                                           | Arch                                                       |
|-------------------|---------------------------|-----------------------------------------------------------|------------------------------------------------------------|
| v0.31.8           | v1.31.8                   | registry.k8s.io/scheduler-plugins/kube-scheduler:v0.31.8  | linux/amd64<br>linux/arm64<br>linux/s390x<br>linux/ppc64le |
| v0.30.12          | v1.30.12                  | registry.k8s.io/scheduler-plugins/kube-scheduler:v0.30.12 | linux/amd64<br>linux/arm64<br>linux/s390x<br>linux/ppc64le |
| v0.29.7           | v1.29.7                   | registry.k8s.io/scheduler-plugins/kube-scheduler:v0.29.7  | linux/amd64<br>linux/arm64<br>linux/s390x<br>linux/ppc64le |

| Controller | Compiled With k8s Version | Container Image                                       | Arch                                                       |
|------------|---------------------------|-------------------------------------------------------|------------------------------------------------------------|
| v0.31.8    | v1.31.8                   | registry.k8s.io/scheduler-plugins/controller:v0.31.8  | linux/amd64<br>linux/arm64<br>linux/s390x<br>linux/ppc64le |
| v0.30.12   | v1.30.12                  | registry.k8s.io/scheduler-plugins/controller:v0.30.12 | linux/amd64<br>linux/arm64<br>linux/s390x<br>linux/ppc64le |
| v0.29.7    | v1.29.7                   | registry.k8s.io/scheduler-plugins/controller:v0.29.7  | linux/amd64<br>linux/arm64<br>linux/s390x<br>linux/ppc64le |

<details>
<summary>Older releases</summary>

| Scheduler Plugins | Compiled With k8s Version | Container Image                                           | Arch                       |
|-------------------|---------------------------|-----------------------------------------------------------|----------------------------|
| v0.28.9           | v1.28.9                   | registry.k8s.io/scheduler-plugins/kube-scheduler:v0.28.9  | linux/amd64<br>linux/arm64 |
| v0.27.8           | v1.27.8                   | registry.k8s.io/scheduler-plugins/kube-scheduler:v0.27.8  | linux/amd64<br>linux/arm64 |
| v0.26.7           | v1.26.7                   | registry.k8s.io/scheduler-plugins/kube-scheduler:v0.26.7  | linux/amd64<br>linux/arm64 |
| v0.25.12          | v1.25.12                  | registry.k8s.io/scheduler-plugins/kube-scheduler:v0.25.12 | linux/amd64<br>linux/arm64 |
| v0.24.9           | v1.24.9                   | registry.k8s.io/scheduler-plugins/kube-scheduler:v0.24.9  | linux/amd64<br>linux/arm64 |
| v0.23.10          | v1.23.10                  | registry.k8s.io/scheduler-plugins/kube-scheduler:v0.23.10 | linux/amd64<br>linux/arm64 |
| v0.22.6           | v1.22.6                   | registry.k8s.io/scheduler-plugins/kube-scheduler:v0.22.6  | linux/amd64<br>linux/arm64 |
| v0.21.6           | v1.21.6                   | registry.k8s.io/scheduler-plugins/kube-scheduler:v0.21.6  | linux/amd64<br>linux/arm64 |
| v0.20.10          | v1.20.10                  | registry.k8s.io/scheduler-plugins/kube-scheduler:v0.20.10 | linux/amd64<br>linux/arm64 |
| v0.19.9           | v1.19.9                   | registry.k8s.io/scheduler-plugins/kube-scheduler:v0.19.9  | linux/amd64<br>linux/arm64 |
| v0.19.8           | v1.19.8                   | registry.k8s.io/scheduler-plugins/kube-scheduler:v0.19.8  | linux/amd64<br>linux/arm64 |
| v0.18.9           | v1.18.9                   | registry.k8s.io/scheduler-plugins/kube-scheduler:v0.18.9  | linux/amd64                |

| Controller | Compiled With k8s Version | Container Image                                       | Arch                       |
|------------|---------------------------|-------------------------------------------------------|----------------------------|
| v0.28.9    | v1.28.9                   | registry.k8s.io/scheduler-plugins/controller:v0.28.9  | linux/amd64<br>linux/arm64 |
| v0.27.8    | v1.27.8                   | registry.k8s.io/scheduler-plugins/controller:v0.27.8  | linux/amd64<br>linux/arm64 |
| v0.26.7    | v1.26.7                   | registry.k8s.io/scheduler-plugins/controller:v0.26.7  | linux/amd64<br>linux/arm64 |
| v0.25.12   | v1.25.12                  | registry.k8s.io/scheduler-plugins/controller:v0.25.12 | linux/amd64<br>linux/arm64 |
| v0.24.9    | v1.24.9                   | registry.k8s.io/scheduler-plugins/controller:v0.24.9  | linux/amd64<br>linux/arm64 |
| v0.23.10   | v1.23.10                  | registry.k8s.io/scheduler-plugins/controller:v0.23.10 | linux/amd64<br>linux/arm64 |
| v0.22.6    | v1.22.6                   | registry.k8s.io/scheduler-plugins/controller:v0.22.6  | linux/amd64<br>linux/arm64 |
| v0.21.6    | v1.21.6                   | registry.k8s.io/scheduler-plugins/controller:v0.21.6  | linux/amd64<br>linux/arm64 |
| v0.20.10   | v1.20.10                  | registry.k8s.io/scheduler-plugins/controller:v0.20.10 | linux/amd64<br>linux/arm64 |
| v0.19.9    | v1.19.9                   | registry.k8s.io/scheduler-plugins/controller:v0.19.9  | linux/amd64<br>linux/arm64 |
| v0.19.8    | v1.19.8                   | registry.k8s.io/scheduler-plugins/controller:v0.19.8  | linux/amd64<br>linux/arm64 |

</details>

## Community, discussion, contribution, and support

Learn how to engage with the Kubernetes community on the [community page](http://kubernetes.io/community/).

You can reach the maintainers of this project at:

- [Slack](https://kubernetes.slack.com/messages/sig-scheduling)
- [Mailing List](https://groups.google.com/forum/#!forum/kubernetes-sig-scheduling)

You can find an [instruction how to build and run out-of-tree plugin here](doc/develop.md) .

### Code of conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](code-of-conduct.md).
