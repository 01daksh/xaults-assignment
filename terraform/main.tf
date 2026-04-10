locals {
  manifests = split("\n---\n", file("../k8s/manifests.yaml"))
}
//splitting manifests into individual resources
// otherwise it will break


resource "kubernetes_manifest" "xaults" {
  for_each = {
    for idx, m in local.manifests :
    idx => m
  }

  manifest = yamldecode(each.value)
}