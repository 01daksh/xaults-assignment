locals {
  documents = [
    for doc in split("\n---\n", file("../k8s/manifests.yaml")) :
    yamldecode(doc)
    if trimspace(doc) != ""
  ]

  manifests = {
    for doc in local.documents :
    "${doc.kind}/${doc.metadata.name}" => doc
  }

  namespaces = {
    for key, doc in local.manifests :
    key => doc
    if doc.kind == "Namespace"
  }

  config_objects = {
    for key, doc in local.manifests :
    key => doc
    if contains(["Secret", "ConfigMap"], doc.kind)
  }

  services = {
    for key, doc in local.manifests :
    key => doc
    if doc.kind == "Service"
  }

  deployments = {
    for key, doc in local.manifests :
    key => doc
    if doc.kind == "Deployment"
  }

  ingresses = {
    for key, doc in local.manifests :
    key => doc
    if doc.kind == "Ingress"
  }
}

resource "kubernetes_manifest" "namespaces" {
  for_each = local.namespaces

  manifest = each.value
}

resource "kubernetes_manifest" "config_objects" {
  for_each = local.config_objects

  manifest   = each.value
  depends_on = [kubernetes_manifest.namespaces]
}

resource "kubernetes_manifest" "services" {
  for_each = local.services

  manifest   = each.value
  depends_on = [kubernetes_manifest.namespaces]
}

resource "kubernetes_manifest" "deployments" {
  for_each = local.deployments

  manifest = each.value
  depends_on = [
    kubernetes_manifest.namespaces,
    kubernetes_manifest.config_objects,
    kubernetes_manifest.services,
  ]
}

resource "kubernetes_manifest" "ingresses" {
  for_each = local.ingresses

  manifest = each.value
  depends_on = [
    kubernetes_manifest.namespaces,
    kubernetes_manifest.services,
    kubernetes_manifest.deployments,
  ]
}
