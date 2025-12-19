#!/bin/bash
# Quick demo MySQL deployment script for FinTrack GKE

set -e

echo "=== Deploying MySQL to Kubernetes ==="
echo ""

# Apply MySQL manifests
echo "1. Creating MySQL Secret..."
kubectl apply -f k8s/mysql-secret.yaml

echo ""
echo "2. Deploying MySQL..."
kubectl apply -f k8s/mysql-deployment.yaml

echo ""
echo "3. Creating MySQL Service..."
kubectl apply -f k8s/mysql-service.yaml

echo ""
echo "4. Waiting for MySQL to be ready..."
kubectl wait --for=condition=ready pod -l app=mysql --timeout=120s

echo ""
echo "5. Updating backend deployment..."
kubectl apply -f k8s/deployment.yaml

echo ""
echo "6. Restarting backend pods to pick up new DB config..."
kubectl rollout restart deployment/fintrack-backend

echo ""
echo "=== Deployment Complete ==="
echo ""
echo "Waiting for backend pods to be ready..."
kubectl wait --for=condition=ready pod -l app=fintrack-backend --timeout=120s

echo ""
echo "=== Verification ==="
echo ""
echo "MySQL Pod Status:"
kubectl get pods -l app=mysql

echo ""
echo "Backend Pod Status:"
kubectl get pods -l app=fintrack-backend

echo ""
echo "Service Status:"
kubectl get svc mysql
kubectl get svc fintrack-backend-service

echo ""
EXTERNAL_IP=$(kubectl get service fintrack-backend-service -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null)
if [ -n "$EXTERNAL_IP" ] && [ "$EXTERNAL_IP" != "null" ]; then
  echo "Testing /health endpoint:"
  curl -s http://$EXTERNAL_IP/health
  echo ""
  echo ""
  echo "Testing /health/ready endpoint:"
  curl -s http://$EXTERNAL_IP/health/ready
  echo ""
else
  echo "LoadBalancer IP pending. Check with:"
  echo "  kubectl get svc fintrack-backend-service"
fi
