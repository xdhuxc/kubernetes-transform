package pkg

var (
	DefaultToken string = "xdhuxc"

	// kubernetes 资源操作
	KubernetesResourceActionInclude string = "include"
	KubernetesResourceActionExclude string = "exclude"

	// 同名资源创建策略
	KubernetesResourcePolicySkip  string = "skip"
	KubernetesResourcePolicyMerge string = "merge"

	// Kubernetes 资源标签操作类型
	KubernetesResourceLabelOperationUpdate string = "update"
	KubernetesResourceLabelOperationDelete string = "delete"
	KubernetesResourceLabelOperationMerge  string = "merge"

	// 请求类型
	RequestTypeTransform string = "transform"
	RequestTypeSave      string = "save"
	RequestTypeRestore   string = "restore"

	// kubernetes 资源类型和 API 版本
	KubernetesResourceDeployment           string = "Deployment"
	KubernetesResourceDeploymentAPIVersion string = "apps/v1"

	KubernetesResourceService           string = "Service"
	KubernetesResourceServiceAPIVersion string = "v1"

	KubernetesResourceIngress           string = "Ingress"
	KubernetesResourceIngressAPIVersion string = "v1beta1"

	KubernetesResourceConfigMap           string = "ConfigMap"
	KubernetesResourceConfigMapAPIVersion string = "v1"

	KubernetesResourceCronJob           string = "CronJob"
	KubernetesResourceCronJobAPIVersion string = "batch/v1beta1"

	KubernetesResourceSecret           string = "Secret"
	KubernetesResourceSecretAPIVersion string = "v1"

	KubernetesResourceNamespace           string = "Namespace"
	KubernetesResourceNamespaceAPIVersion string = "v1"
)
