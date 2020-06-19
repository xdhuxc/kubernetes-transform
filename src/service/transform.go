package service

import (
	"fmt"

	"github.com/asaskevich/govalidator"
	mapset "github.com/deckarep/golang-set"
	appv1 "k8s.io/api/apps/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	"github.com/xdhuxc/kubernetes-transform/src/config"
	"github.com/xdhuxc/kubernetes-transform/src/model"
	"github.com/xdhuxc/kubernetes-transform/src/pkg"
	"github.com/xdhuxc/kubernetes-transform/src/util"
)

type transformService struct {
	sc  model.Cluster         // source cluster
	tc  model.Cluster         // target cluster
	skc *kubernetes.Clientset // source kubernetes cluster client
	tkc *kubernetes.Clientset // target kubernetes cluster client
	cnf config.Config
}

func newTransformService(cnf config.Config, skc *kubernetes.Clientset, tkc *kubernetes.Clientset) *transformService {
	return &transformService{
		sc:  cnf.Source,
		tc:  cnf.Target,
		skc: skc,
		tkc: tkc,
		cnf: cnf,
	}
}

func (ts *transformService) Check() error {
	if ts.skc == nil {
		return fmt.Errorf("源集群 Kubernetes 客户端未初始化。")
	}
	if ts.tkc == nil {
		return fmt.Errorf("目标集群 Kubernetes 客户端未初始化。")
	}

	return nil
}

func (ts *transformService) Transform(tr *model.TransformRequest) error {
	_, nss, err := Namespaces(ts.skc)
	if err != nil {
		return err
	}

	namespaces := mapset.NewSetFromSlice(util.Convert2Interfaces(nss))
	// Namespace Request Set
	nrs := mapset.NewSetFromSlice(util.Convert2Interfaces(ts.cnf.Namespace.Namespaces))
	kinds := mapset.NewSetFromSlice(util.Convert2Interfaces(ts.cnf.Resource.Kinds))
	// Resource Request Set
	rrs := mapset.NewSetFromSlice(util.Convert2Interfaces(ts.cnf.Resource.Resources))

	if ts.cnf.Namespace.Action == pkg.KubernetesResourceActionInclude {
		// Namespace Inclusion Set
		nis := namespaces.Intersect(nrs)
		if nis.Cardinality() > 0 {
			if ts.cnf.Resource.Action == pkg.KubernetesResourceActionInclude {
				// Resource Inclusion Set
				ris := kinds.Intersect(rrs)
				if ris.Cardinality() > 0 {
					ts.Execute(nis, ris)
				} else {
					return fmt.Errorf("there is no intersection between the requested resources and resources that has been coded currently")
				}
			} else if ts.cnf.Resource.Action == pkg.KubernetesResourceActionExclude {
				// Resource Inclusion Set
				ris := kinds.Difference(rrs)
				if ris.Cardinality() > 0 {
					ts.Execute(nis, ris)
				} else {
					return fmt.Errorf("the difference between the requested resources and the qctual resources is an empty set")
				}
			} else {
				return fmt.Errorf("the action of resource request is invalid")
			}
		} else {
			return fmt.Errorf("请求的命名空间和集群实际具有的命名空间之间没有交集。")
		}
	} else if ts.cnf.Namespace.Action == pkg.KubernetesResourceActionExclude {
		// Namespace Inclusion Set
		nis := namespaces.Difference(nrs)
		if nis.Cardinality() > 0 {
			if ts.cnf.Resource.Action == pkg.KubernetesResourceActionInclude {
				// Resource Inclusion Set
				ris := kinds.Intersect(rrs)
				if ris.Cardinality() > 0 {
					ts.Execute(nis, ris)
				} else {
					return fmt.Errorf("there is no intersection between the requested resources and resources that has been coded currently")
				}
			} else if ts.cnf.Resource.Action == pkg.KubernetesResourceActionExclude {
				// Resource Inclusion Set
				ris := kinds.Difference(rrs)
				if ris.Cardinality() > 0 {
					ts.Execute(nis, ris)
				} else {
					return fmt.Errorf("the difference between the resources of actual and the requested resources is an empty set")
				}
			} else {
				return fmt.Errorf("the action of resource request is invalid")
			}
		} else {
			return fmt.Errorf("the difference between the actual namespaces of the cluster and the requested namespaces is an empty set")
		}
	} else {
		return fmt.Errorf("the action of namespace request is invalid")
	}

	return nil
}

func (ts *transformService) Deployment(namespace string) error {
	sdc := ts.skc.AppsV1().Deployments(namespace)
	tdc := ts.tkc.AppsV1().Deployments(namespace)

	// 处理 continue 参数
	token := pkg.DefaultToken
	var options metav1.ListOptions
	getOptions := metav1.GetOptions{}

	for {
		if token != pkg.DefaultToken {
			options = metav1.ListOptions{
				Continue: token,
			}
		} else {
			options = metav1.ListOptions{}
		}

		deploymentList, err := sdc.List(options)
		if err != nil {
			return err
		}
		for _, item := range deploymentList.Items {
			_, err := tdc.Get(item.Name, getOptions)
			if err == nil {
				klog.Infof("the deployment %s in source cluster %s has already existed in target cluster %s", item.Name, ts.sc.Name, ts.tc.Name)
				continue
			}

			// 标签处理
			deployment := &appv1.Deployment{
				TypeMeta: metav1.TypeMeta{
					Kind:       item.TypeMeta.Kind,
					APIVersion: item.TypeMeta.APIVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:        item.ObjectMeta.Name,
					Namespace:   item.ObjectMeta.Namespace,
					Labels:      item.ObjectMeta.Labels,
					Annotations: item.ObjectMeta.Annotations,
				},
				Spec: appv1.DeploymentSpec{
					Replicas: item.Spec.Replicas,
					Selector: item.Spec.Selector,
					Template: v1.PodTemplateSpec{
						ObjectMeta: item.Spec.Template.ObjectMeta,
						Spec:       item.Spec.Template.Spec,
					},
				},
			}

			_, err = tdc.Create(deployment)
			if err != nil {
				klog.Errorln(err)
				continue
			}
			klog.Infof("created deployment %s in target cluster %s", deployment.Name, ts.tc.Name)
		}

		token = deploymentList.Continue
		if govalidator.IsNull(token) {
			break
		}
	}

	return nil
}

func (ts *transformService) Service(namespace string) error {
	// source/target service client
	ssc := ts.skc.CoreV1().Services(namespace)
	tsc := ts.tkc.CoreV1().Services(namespace)

	// 处理 continue 参数
	token := pkg.DefaultToken
	var options metav1.ListOptions
	getOptions := metav1.GetOptions{}
	for {
		if token != pkg.DefaultToken {
			options = metav1.ListOptions{
				Continue: token,
			}
		} else {
			options = metav1.ListOptions{}
		}

		serviceList, err := ssc.List(options)
		if err != nil {
			return err
		}
		for _, item := range serviceList.Items {
			_, err := tsc.Get(item.Name, getOptions)
			if err == nil {
				klog.Infof("the service %s in source cluster %s has already existed in target cluster %s", item.Name, ts.sc.Name, ts.tc.Name)
				continue
			}

			// 标签处理
			service := &v1.Service{
				TypeMeta: metav1.TypeMeta{
					Kind:       item.TypeMeta.Kind,
					APIVersion: item.TypeMeta.APIVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:        item.ObjectMeta.Name,
					Namespace:   item.ObjectMeta.Namespace,
					Labels:      item.ObjectMeta.Labels,
					Annotations: item.ObjectMeta.Annotations,
				},
				Spec: v1.ServiceSpec{
					Ports:    item.Spec.Ports,
					Type:     item.Spec.Type,
					Selector: item.Spec.Selector,
				},
			}

			_, err = tsc.Create(service)
			if err != nil {
				klog.Errorln(err)
				continue
			}
			klog.Infof("created service %s in target cluster %s", service.Name, ts.tc.Name)
		}

		token = serviceList.Continue
		if govalidator.IsNull(token) {
			break
		}
	}

	return nil
}

func (ts *transformService) Ingress(namespace string) error {
	// source ingress client
	sic := ts.skc.ExtensionsV1beta1().Ingresses(namespace)
	// target ingress client
	tic := ts.tkc.ExtensionsV1beta1().Ingresses(namespace)

	// 处理 continue 参数
	token := pkg.DefaultToken
	var options metav1.ListOptions
	getOptions := metav1.GetOptions{}
	for {
		if token != pkg.DefaultToken {
			options = metav1.ListOptions{
				Continue: token,
			}
		} else {
			options = metav1.ListOptions{}
		}

		ingressList, err := sic.List(options)
		if err != nil {
			return err
		}
		for _, item := range ingressList.Items {
			_, err := tic.Get(item.Name, getOptions)
			if err == nil {
				klog.Infof("the ingress %s in source cluster %s has already existed in target cluster %s", item.Name, ts.sc.Name, ts.tc.Name)
				continue
			}

			// 标签处理
			ingress := &v1beta1.Ingress{
				TypeMeta: metav1.TypeMeta{
					Kind:       item.TypeMeta.Kind,
					APIVersion: item.TypeMeta.APIVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:        item.ObjectMeta.Name,
					Namespace:   item.ObjectMeta.Namespace,
					Labels:      item.ObjectMeta.Labels,
					Annotations: item.ObjectMeta.Annotations,
				},
				Spec: v1beta1.IngressSpec{
					Backend: item.Spec.Backend,
					TLS:     item.Spec.TLS,
					Rules:   item.Spec.Rules,
				},
			}
			_, err = tic.Create(ingress)
			if err != nil {
				klog.Errorln(err)
				continue
			}
			klog.Infof("created ingress %s in target cluster %s", ingress.Name, ts.tc.Name)
		}

		token = ingressList.Continue
		if govalidator.IsNull(token) {
			break
		}
	}

	return nil
}

func (ts *transformService) ConfigMap(namespace string) error {
	// source ConfigMap client
	scc := ts.skc.CoreV1().ConfigMaps(namespace)
	// target ConfigMap client
	tcc := ts.tkc.CoreV1().ConfigMaps(namespace)

	// 处理 continue 参数
	token := pkg.DefaultToken
	var options metav1.ListOptions
	getOptions := metav1.GetOptions{}
	for {
		if token != pkg.DefaultToken {
			options = metav1.ListOptions{
				Continue: token,
			}
		} else {
			options = metav1.ListOptions{}
		}

		configMapList, err := scc.List(options)
		if err != nil {
			return err
		}
		for _, item := range configMapList.Items {
			_, err := tcc.Get(item.Name, getOptions)
			if err == nil {
				klog.Infof("the configMap %s in source cluster %s has already existed in target cluster %s", item.Name, ts.sc.Name, ts.tc.Name)
				continue
			}

			configMap := &v1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					Kind:       item.TypeMeta.Kind,
					APIVersion: item.TypeMeta.APIVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:        item.ObjectMeta.Name,
					Namespace:   item.ObjectMeta.Namespace,
					Labels:      item.ObjectMeta.Labels,
					Annotations: item.ObjectMeta.Annotations,
				},
				Data:       item.Data,
				BinaryData: item.BinaryData,
			}
			_, err = tcc.Create(configMap)
			if err != nil {
				klog.Errorln(err)
				continue
			}
			klog.Infof("created configMap %s in target cluster %s", configMap.Name, ts.tc.Name)
		}

		token = configMapList.Continue
		if govalidator.IsNull(token) {
			break
		}
	}

	return nil
}

func (ts *transformService) CronJob(namespace string) error {
	// source CronJob client
	scc := ts.skc.BatchV1beta1().CronJobs(namespace)
	// target CronJob client
	tcc := ts.tkc.BatchV1beta1().CronJobs(namespace)

	// 处理 continue 参数
	token := pkg.DefaultToken
	var options metav1.ListOptions
	getOptions := metav1.GetOptions{}
	for {
		if token != pkg.DefaultToken {
			options = metav1.ListOptions{
				Continue: token,
			}
		} else {
			options = metav1.ListOptions{}
		}

		cronJobList, err := scc.List(options)
		if err != nil {
			return err
		}
		for _, item := range cronJobList.Items {
			_, err := tcc.Get(item.Name, getOptions)
			if err == nil {
				klog.Infof("the cronJob %s in source cluster %s has already existed in target cluster %s", item.Name, ts.sc.Name, ts.tc.Name)
				continue
			}

			cronJob := &batchv1beta1.CronJob{
				TypeMeta: metav1.TypeMeta{
					Kind:       item.TypeMeta.Kind,
					APIVersion: item.TypeMeta.APIVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:        item.ObjectMeta.Name,
					Namespace:   item.ObjectMeta.Namespace,
					Labels:      item.ObjectMeta.Labels,
					Annotations: item.ObjectMeta.Annotations,
				},
				Spec: batchv1beta1.CronJobSpec{
					Schedule:                   item.Spec.Schedule,
					StartingDeadlineSeconds:    item.Spec.StartingDeadlineSeconds,
					SuccessfulJobsHistoryLimit: item.Spec.SuccessfulJobsHistoryLimit,
					FailedJobsHistoryLimit:     item.Spec.FailedJobsHistoryLimit,
					ConcurrencyPolicy:          item.Spec.ConcurrencyPolicy,
					JobTemplate:                item.Spec.JobTemplate,
				},
			}

			_, err = tcc.Create(cronJob)
			if err != nil {
				klog.Errorln(err)
				continue
			}
			klog.Infof("created cronJob %s in target cluster %s", cronJob.Name, ts.tc.Name)
		}

		token = cronJobList.Continue
		if govalidator.IsNull(token) {
			break
		}
	}

	return nil
}

func (ts *transformService) Namespace() error {
	// source namespace client
	snc := ts.skc.CoreV1().Namespaces()
	// target namespace client
	tnc := ts.tkc.CoreV1().Namespaces()

	// 处理 continue 参数
	token := pkg.DefaultToken
	var options metav1.ListOptions
	getOptions := metav1.GetOptions{}
	for {
		if token != pkg.DefaultToken {
			options = metav1.ListOptions{
				Continue: token,
			}
		} else {
			options = metav1.ListOptions{}
		}

		namespaceList, err := snc.List(options)
		if err != nil {
			return err
		}
		for _, item := range namespaceList.Items {
			_, err := tnc.Get(item.Name, getOptions)
			if err == nil {
				klog.Infof("the namespace %s in source cluster %s has already existed in target cluster %s", item.Name, ts.sc.Name, ts.tc.Name)
				continue
			}
			namespace := &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:   item.ObjectMeta.Name,
					Labels: item.ObjectMeta.Labels,
				},
			}

			_, err = tnc.Create(namespace)
			if err != nil {
				klog.Errorln(err)
				continue
			}
			klog.Infof("created cronJob %s in target cluster %s", namespace.Name, ts.tc.Name)
		}

		token = namespaceList.Continue
		if govalidator.IsNull(token) {
			break
		}
	}

	return nil
}

func (ts *transformService) Secret(namespace string) error {
	// source secret client
	ssc := ts.skc.CoreV1().Secrets(namespace)
	// target secret client
	tsc := ts.tkc.CoreV1().Secrets(namespace)

	// 处理 continue 参数
	token := pkg.DefaultToken
	var options metav1.ListOptions
	getOptions := metav1.GetOptions{}
	for {
		if token != pkg.DefaultToken {
			options = metav1.ListOptions{
				Continue: token,
			}
		} else {
			options = metav1.ListOptions{}
		}

		secretList, err := ssc.List(options)
		if err != nil {
			return err
		}
		for _, item := range secretList.Items {
			_, err := tsc.Get(item.Name, getOptions)
			if err == nil {
				klog.Infof("the secret %s in source cluster %s has already existed in target cluster %s", item.Name, ts.sc.Name, ts.tc.Name)
				continue
			}

			secret := &v1.Secret{
				TypeMeta: metav1.TypeMeta{
					Kind:       pkg.KubernetesResourceSecret,
					APIVersion: pkg.KubernetesResourceSecretAPIVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:        item.ObjectMeta.Name,
					Namespace:   item.ObjectMeta.Namespace,
					Labels:      item.ObjectMeta.Labels,
					Annotations: item.ObjectMeta.Annotations,
				},
				Type:       item.Type,
				Data:       item.Data,
				StringData: item.StringData,
			}

			_, err = tsc.Create(secret)
			if err != nil {
				klog.Errorln(err)
				continue
			}
			klog.Infof("created cronJob %s in target cluster %s", secret.Name, ts.tc.Name)
		}

		token = secretList.Continue
		if govalidator.IsNull(token) {
			break
		}
	}

	return nil
}

func (ts *transformService) Execute(nis mapset.Set, ris mapset.Set) {
	namespaces := util.Convert2Strings(nis.ToSlice())
	var err error
	for _, namespace := range namespaces {
		if ris.Contains(pkg.KubernetesResourceDeployment) {
			err = ts.Deployment(namespace)
			klog.Errorln(err)
		}
		if ris.Contains(pkg.KubernetesResourceService) {
			err = ts.Service(namespace)
			klog.Errorln(err)
		}
		if ris.Contains(pkg.KubernetesResourceIngress) {
			err = ts.Ingress(namespace)
			klog.Errorln(err)
		}
		if ris.Contains(pkg.KubernetesResourceConfigMap) {
			err = ts.ConfigMap(namespace)
			klog.Errorln(err)
		}
		if ris.Contains(pkg.KubernetesResourceCronJob) {
			err = ts.CronJob(namespace)
			klog.Errorln(err)
		}
		if ris.Contains(pkg.KubernetesResourceSecret) {
			err = ts.Secret(namespace)
			klog.Errorln(err)
		}
	}

	return
}
