#!/usr/bin/env bash


kubectl create -f app1/k8s/app1.yaml
kubectl create -f app1/k8s/service1.yaml

kubectl create -f app2/k8s/app2.yaml
kubectl create -f app2/k8s/service2.yaml
