# Kubernetes Guide for MailRaven

This guide teaches you Kubernetes concepts and walks through deploying MailRaven locally.

---

## Part 1: Kubernetes Concepts (The 5-Minute Version)

### What is Kubernetes?

Kubernetes (K8s) is a container orchestrator. Think of it as a **smart supervisor** that:
- Runs your containers (called **Pods**)
- Restarts them if they crash (**self-healing**)
- Scales them up/down based on load (**autoscaling**)
- Routes network traffic to them (**Services**)
- Manages secrets and config separately from code

### Key Concepts

| Concept | What it is | MailRaven Example |
|---------|-----------|-------------------|
| **Pod** | Smallest unit — one or more containers running together | One MailRaven backend instance |
| **Deployment** | Manages multiple identical Pods (replicas) | "Run 2 MailRaven backends" |
| **StatefulSet** | Like Deployment but for stateful services (databases) | PostgreSQL, Redis, MinIO |
| **Service** | Stable network endpoint that routes to Pods | `mailraven-smtp` on port 25 |
| **Ingress** | HTTP routing from outside the cluster | `mail.example.com` → backend |
| **ConfigMap** | Non-secret configuration (domain name, ports) | `MAILRAVEN_DOMAIN=mail.example.com` |
| **Secret** | Sensitive values (passwords, keys) | `MAILRAVEN_JWT_SECRET`, DB password |
| **PVC** | Persistent storage that survives Pod restarts | PostgreSQL data, MinIO blobs |
| **Namespace** | Logical isolation (like folders) | `mailraven` namespace for all our stuff |
| **HPA** | Horizontal Pod Autoscaler — scales replicas | Scale backend 1→10 based on CPU |
| **KEDA** | Advanced autoscaler — can scale to zero | Scale backend 0→10 based on SMTP connections |

### How it fits together

```
You (kubectl) → Kubernetes API → Scheduler → Nodes (your machine)
                                                 ↓
                                    ┌─────────────────────┐
                                    │ Pod: mailraven-backend│
                                    │ Pod: mailraven-backend│ (2 replicas)
                                    │ Pod: postgres         │ (StatefulSet)
                                    │ Pod: redis            │ (StatefulSet)
                                    │ Pod: nats             │ (StatefulSet)
                                    │ Pod: minio            │ (StatefulSet)
                                    └─────────────────────┘
```

### The Lifecycle

1. You write YAML manifests describing your desired state
2. `kubectl apply` sends them to the cluster
3. Kubernetes makes reality match your description
4. If a pod crashes → K8s restarts it automatically
5. If load increases → HPA/KEDA adds more pods
6. If load drops → pods scale down (even to zero with KEDA)

---

## Part 2: Local Setup with Docker Desktop

### Step 1: Enable Kubernetes

1. Open **Docker Desktop**
2. Go to **Settings** (gear icon) → **Kubernetes**
3. Check **"Enable Kubernetes"**
4. Click **"Apply & restart"**
5. Wait 2-3 minutes for the cluster to start

Verify it works:
```bash
kubectl cluster-info
kubectl get nodes
```

You should see one node (`docker-desktop`) in `Ready` state.

### Step 2: Install kubectl (if not already)

```bash
# macOS with Homebrew
brew install kubectl

# Verify
kubectl version --client
```

### Step 3: Build the MailRaven Docker Image Locally

Since we're running locally, we build the image and it's immediately available to K8s:

```bash
cd /path/to/MailRaven-Server

# Build the backend image
docker build -t mailraven-server:local -f build/Dockerfile .

# Build the frontend image
docker build -t mailraven-frontend:local -f build/Dockerfile.frontend .
```

### Step 4: Update Manifests for Local Use

The manifests in `deployment/kubernetes/` reference `ghcr.io/...` images. For local testing, we need to use our local images.

Create a local override:
```bash
# Create a local kustomization overlay
mkdir -p deployment/kubernetes/local
```

We'll create this file below (Step 5).

### Step 5: Deploy to Local Kubernetes

```bash
# Create the namespace
kubectl create namespace mailraven

# Apply secrets (edit these first!)
kubectl apply -f deployment/kubernetes/base/secret.yaml

# Apply the full stack
kubectl apply -k deployment/kubernetes/

# Watch pods come up
kubectl -n mailraven get pods -w
```

Expected output (after ~1 minute):
```
NAME                         READY   STATUS    RESTARTS   AGE
mailraven-backend-xxx-yyy    1/1     Running   0          30s
mailraven-backend-xxx-zzz    1/1     Running   0          30s
postgres-0                   1/1     Running   0          45s
redis-0                      1/1     Running   0          40s
nats-0                       1/1     Running   0          40s
minio-0                      1/1     Running   0          40s
```

### Step 6: Access the Services

```bash
# Port-forward the API to localhost:8080
kubectl -n mailraven port-forward svc/mailraven-http 8080:8080

# Port-forward SMTP to localhost:2525 (can't use 25 without sudo)
kubectl -n mailraven port-forward svc/mailraven-smtp 2525:25

# Port-forward the frontend
kubectl -n mailraven port-forward svc/mailraven-frontend 3000:80
```

Now you can:
- API: http://localhost:8080/health
- Frontend: http://localhost:3000
- SMTP: `telnet localhost 2525`

---

## Part 3: Useful kubectl Commands

```bash
# See all pods
kubectl -n mailraven get pods

# See logs for a pod
kubectl -n mailraven logs deployment/mailraven-backend

# Follow logs in real-time
kubectl -n mailraven logs -f deployment/mailraven-backend

# Get into a pod's shell
kubectl -n mailraven exec -it deployment/mailraven-backend -- sh

# See resource usage
kubectl -n mailraven top pods

# Describe a pod (troubleshooting)
kubectl -n mailraven describe pod <pod-name>

# Scale manually
kubectl -n mailraven scale deployment mailraven-backend --replicas=3

# Delete everything
kubectl delete namespace mailraven
```

---

## Part 4: Testing Performance Locally

### Test API Latency

```bash
# Install hey (HTTP load testing tool)
brew install hey

# Warmup + test (100 concurrent, 1000 requests)
hey -n 1000 -c 100 http://localhost:8080/health
```

### Test SMTP Throughput

```bash
# Install swaks (SMTP testing tool)
brew install swaks

# Send a test email
swaks --to user@yourdomain.com --from test@test.com \
  --server localhost:2525 --body "Test message"

# Load test with multiple messages
for i in $(seq 1 100); do
  swaks --to user@yourdomain.com --from "test$i@test.com" \
    --server localhost:2525 --body "Message $i" &
done
wait
```

### Monitor Resource Usage

```bash
# Watch pod CPU/memory in real-time
watch kubectl -n mailraven top pods

# Check metrics endpoint
curl http://localhost:8080/metrics
```

---

## Part 5: How Scaling Works

### Manual Scaling
```bash
kubectl -n mailraven scale deployment mailraven-backend --replicas=4
```

### Autoscaling (HPA)
```bash
# Create HPA: scale 2-10 pods based on 70% CPU
kubectl -n mailraven autoscale deployment mailraven-backend \
  --min=2 --max=10 --cpu-percent=70
```

### Scale-to-Zero (KEDA)

KEDA is an add-on that watches custom metrics and scales to zero:

```bash
# Install KEDA
kubectl apply --server-side \
  -f https://github.com/kedacore/keda/releases/download/v2.13.0/keda-2.13.0.yaml

# Apply our ScaledObject
kubectl apply -f deployment/kubernetes/keda/scaledobject.yaml
```

With KEDA active:
- No SMTP connections for 5 min → pods scale to 0
- New SMTP connection arrives → pod scales up in ~10 seconds
- Queue depth > 5 → more pods added

---

## Part 6: Troubleshooting

### Pods stuck in "Pending"
```bash
kubectl -n mailraven describe pod <pod-name>
# Look for "Events" section — usually a resource or PVC issue
```

### Pods in "CrashLoopBackOff"
```bash
kubectl -n mailraven logs <pod-name> --previous
# Shows logs from the crashed container
```

### Can't connect to services
```bash
# Check service endpoints
kubectl -n mailraven get endpoints

# Check if pods are ready
kubectl -n mailraven get pods -o wide
```

### Database connection refused
```bash
# Check if postgres is running
kubectl -n mailraven get pods -l app=postgres

# Check postgres logs
kubectl -n mailraven logs statefulset/postgres
```

---

## Part 7: Local Development Workflow

The fastest dev loop for K8s:

1. **Change code**
2. **Rebuild image**: `docker build -t mailraven-server:local -f build/Dockerfile .`
3. **Restart deployment**: `kubectl -n mailraven rollout restart deployment mailraven-backend`
4. **Watch logs**: `kubectl -n mailraven logs -f deployment/mailraven-backend`

For faster iteration, use Docker Compose during development and K8s only for final integration testing.

---

## Migration from SQLite to PostgreSQL

If you started with SQLite and want to migrate:

1. Export data from SQLite:
   ```bash
   mailraven-cli export --format json > data.json
   ```

2. Update config to use PostgreSQL:
   ```yaml
   storage:
     driver: postgres
     dsn: "postgres://mailraven:password@localhost:5432/mailraven"
   ```

3. Start with PostgreSQL (migrations run automatically):
   ```bash
   mailraven serve --config /etc/mailraven/config.yaml
   ```

4. Import data:
   ```bash
   mailraven-cli import --format json < data.json
   ```

Note: The import/export CLI commands are planned features. For now, direct SQL migration is needed for existing deployments.
