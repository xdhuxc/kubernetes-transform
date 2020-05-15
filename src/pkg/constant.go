package pkg

var (
	DefaultToken = "xdhuxc"

	KubernetesResourceActionInclude = "include"
	KubernetesResourceActionExclude = "exclude"

	// kubernetes 资源类型和 API 版本
	KubernetesResourceDeployment           = "Deployment"
	KubernetesResourceDeploymentAPIVersion = "apps/v1"

	KubernetesResourceService           = "Service"
	KubernetesResourceServiceAPIVersion = "v1"

	KubernetesResourceIngress           = "Ingress"
	KubernetesResourceIngressAPIVersion = "v1beta1"

	KubernetesResourceConfigMap           = "ConfigMap"
	KubernetesResourceConfigMapAPIVersion = "v1"

	KubernetesResourceCronJob           = "CronJob"
	KubernetesResourceCronJobAPIVersion = "batch/v1beta1"

	KubernetesResourceSecret           = "Secret"
	KubernetesResourceSecretAPIVersion = "v1"

	KubernetesResourceNamespace           = "Namespace"
	KubernetesResourceNamespaceAPIVersion = "v1"
)
