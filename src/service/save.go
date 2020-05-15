package service

import (
	"fmt"
	"time"

	"github.com/asaskevich/govalidator"
	mapset "github.com/deckarep/golang-set"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	appv1 "k8s.io/api/apps/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"

	"github.com/xdhuxc/kubernetes-transform/src/config"
	"github.com/xdhuxc/kubernetes-transform/src/model"
	"github.com/xdhuxc/kubernetes-transform/src/pkg"
	"github.com/xdhuxc/kubernetes-transform/src/util"
)

type saveService struct {
	sc  model.Cluster         // source cluster
	skc *kubernetes.Clientset // source kubernetes cluster client
	db  *gorm.DB
	cnf config.Config
}

func newSaveService(cnf config.Config, skc *kubernetes.Clientset, db *gorm.DB) *saveService {
	return &saveService{
		sc:  cnf.Source,
		skc: skc,
		db:  db,
		cnf: cnf,
	}
}

func (ss *saveService) Save() error {
	_, nss, err := Namespaces(ss.skc)
	if err != nil {
		return err
	}

	namespaces := mapset.NewSetFromSlice(util.Convert2Interfaces(nss))
	// Namespace Request Set
	nrs := mapset.NewSetFromSlice(util.Convert2Interfaces(ss.cnf.Namespace.Namespaces))
	kinds := mapset.NewSetFromSlice(util.Convert2Interfaces(ss.cnf.Resource.Kinds))
	// Resource Request Set
	rrs := mapset.NewSetFromSlice(util.Convert2Interfaces(ss.cnf.Resource.Resources))

	if ss.cnf.Namespace.Action == pkg.KubernetesResourceActionInclude {
		// Namespace Inclusion Set
		nis := namespaces.Intersect(nrs)
		if nis.Cardinality() > 0 {
			if ss.cnf.Resource.Action == pkg.KubernetesResourceActionInclude {
				// Resource Inclusion Set
				ris := kinds.Intersect(rrs)
				if ris.Cardinality() > 0 {
					ss.Execute(nis, ris)
				} else {
					return fmt.Errorf("there is no intersection between the requested resources and resources that has been coded currently")
				}
			} else if ss.cnf.Resource.Action == pkg.KubernetesResourceActionExclude {
				// Resource Inclusion Set
				ris := kinds.Difference(rrs)
				if ris.Cardinality() > 0 {
					ss.Execute(nis, ris)
				} else {
					return fmt.Errorf("the difference between the requested resources and the qctual resources is an empty set")
				}
			} else {
				return fmt.Errorf("the action of resource request is invalid")
			}
		} else {
			return fmt.Errorf("there is no intersection between the requested namespaces and the actual namespaces")
		}
	} else if ss.cnf.Namespace.Action == pkg.KubernetesResourceActionExclude {
		// Namespace Inclusion Set
		nis := namespaces.Difference(nrs)
		if nis.Cardinality() > 0 {
			if ss.cnf.Resource.Action == pkg.KubernetesResourceActionInclude {
				// Resource Inclusion Set
				ris := kinds.Intersect(rrs)
				if ris.Cardinality() > 0 {
					ss.Execute(nis, ris)
				} else {
					return fmt.Errorf("there is no intersection between the requested resources and resources that has been coded currently")
				}
			} else if ss.cnf.Resource.Action == pkg.KubernetesResourceActionExclude {
				// Resource Inclusion Set
				ris := kinds.Difference(rrs)
				if ris.Cardinality() > 0 {
					ss.Execute(nis, ris)
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

	if err := ss.afterSave(); err != nil {
		return err
	}

	return nil
}

func (ss *saveService) Deployment(namespace string) error {
	// source Deployment client
	sdc := ss.skc.AppsV1().Deployments(namespace)

	// 处理 continue 参数
	token := pkg.DefaultToken
	var options metav1.ListOptions
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
			var resource model.Resource
			resource.Name = item.ObjectMeta.Name
			resource.Namespace = item.ObjectMeta.Namespace
			resource.Kind = pkg.KubernetesResourceDeployment
			resource.UUID = uuid.New().String()
			resource.CreateTime = time.Now()
			resource.UpdateTime = time.Now()

			deployment := &appv1.Deployment{
				TypeMeta: metav1.TypeMeta{
					Kind:       pkg.KubernetesResourceDeployment,
					APIVersion: pkg.KubernetesResourceDeploymentAPIVersion,
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
					Strategy: item.Spec.Strategy,
					Template: v1.PodTemplateSpec{
						ObjectMeta: item.Spec.Template.ObjectMeta,
						Spec:       item.Spec.Template.Spec,
					},
				},
			}
			resource.Json, resource.Yaml, err = util.Marshal(deployment)
			if err != nil {
				klog.Errorf("Marshal %s %s error : %s", pkg.KubernetesResourceDeployment, resource.Name, err)
				continue
			}

			if err = ss.FirstOrUpdate(resource); err != nil {
				klog.Errorln(err)
				continue
			}
		}

		token = deploymentList.Continue
		if govalidator.IsNull(token) {
			break
		}
	}

	return nil
}

func (ss *saveService) Service(namespace string) error {
	// source Service client
	ssc := ss.skc.CoreV1().Services(namespace)

	// 处理 continue 参数
	token := pkg.DefaultToken
	var options metav1.ListOptions
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
			var resource model.Resource
			resource.Name = item.ObjectMeta.Name
			resource.Namespace = item.ObjectMeta.Namespace
			resource.Kind = pkg.KubernetesResourceService
			resource.UUID = uuid.New().String()
			resource.CreateTime = time.Now()
			resource.UpdateTime = time.Now()

			service := &v1.Service{
				TypeMeta: metav1.TypeMeta{
					Kind:       pkg.KubernetesResourceService,
					APIVersion: pkg.KubernetesResourceServiceAPIVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:        item.ObjectMeta.Name,
					Namespace:   item.ObjectMeta.Namespace,
					Labels:      item.ObjectMeta.Labels,
					Annotations: item.ObjectMeta.Annotations,
				},
				Spec: v1.ServiceSpec{
					Ports:                    item.Spec.Ports,
					Selector:                 item.Spec.Selector,
					Type:                     item.Spec.Type,
					ExternalIPs:              item.Spec.ExternalIPs,
					SessionAffinity:          item.Spec.SessionAffinity,
					LoadBalancerIP:           item.Spec.LoadBalancerIP,
					LoadBalancerSourceRanges: item.Spec.LoadBalancerSourceRanges,
					ExternalName:             item.Spec.ExternalName,
					ExternalTrafficPolicy:    item.Spec.ExternalTrafficPolicy,
					HealthCheckNodePort:      item.Spec.HealthCheckNodePort,
					SessionAffinityConfig:    item.Spec.SessionAffinityConfig,
					IPFamily:                 item.Spec.IPFamily,
					TopologyKeys:             item.Spec.TopologyKeys,
				},
			}
			resource.Json, resource.Yaml, err = util.Marshal(service)
			if err != nil {
				klog.Errorf("Marshal %s %s error : %s", pkg.KubernetesResourceService, resource.Name, err)
				continue
			}

			if err = ss.FirstOrUpdate(resource); err != nil {
				klog.Errorln(err)
				continue
			}
		}

		token = serviceList.Continue
		if govalidator.IsNull(token) {
			break
		}
	}

	return nil
}

func (ss *saveService) Ingress(namespace string) error {
	// source ingress client
	sic := ss.skc.ExtensionsV1beta1().Ingresses(namespace)

	// 处理 continue 参数
	token := pkg.DefaultToken
	var options metav1.ListOptions
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
			var resource model.Resource
			resource.Name = item.ObjectMeta.Name
			resource.Namespace = item.ObjectMeta.Namespace
			resource.Kind = pkg.KubernetesResourceIngress
			resource.UUID = uuid.New().String()
			resource.CreateTime = time.Now()
			resource.UpdateTime = time.Now()

			ingress := &v1beta1.Ingress{
				TypeMeta: metav1.TypeMeta{
					Kind:       pkg.KubernetesResourceIngress,
					APIVersion: pkg.KubernetesResourceIngressAPIVersion,
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
			resource.Json, resource.Yaml, err = util.Marshal(ingress)
			if err != nil {
				klog.Errorf("Marshal %s %s error : %s", pkg.KubernetesResourceIngress, resource.Name, err)
				continue
			}

			if err = ss.FirstOrUpdate(resource); err != nil {
				klog.Errorln(err)
				continue
			}
		}

		token = ingressList.Continue
		if govalidator.IsNull(token) {
			break
		}
	}

	return nil
}

func (ss *saveService) ConfigMap(namespace string) error {
	// source ConfigMap client
	scmc := ss.skc.CoreV1().ConfigMaps(namespace)

	// 处理 continue 参数
	token := pkg.DefaultToken
	var options metav1.ListOptions
	for {
		if token != pkg.DefaultToken {
			options = metav1.ListOptions{
				Continue: token,
			}
		} else {
			options = metav1.ListOptions{}
		}

		configMapList, err := scmc.List(options)
		if err != nil {
			return err
		}
		for _, item := range configMapList.Items {
			var resource model.Resource
			resource.Name = item.ObjectMeta.Name
			resource.Namespace = item.ObjectMeta.Namespace
			resource.Kind = pkg.KubernetesResourceConfigMap
			resource.UUID = uuid.New().String()
			resource.CreateTime = time.Now()
			resource.UpdateTime = time.Now()

			configMap := &v1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					Kind:       pkg.KubernetesResourceConfigMap,
					APIVersion: pkg.KubernetesResourceConfigMapAPIVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:        item.ObjectMeta.Name,
					Namespace:   item.ObjectMeta.Namespace,
					Labels:      item.ObjectMeta.Labels,
					Annotations: item.ObjectMeta.Annotations,
				},
				Data: item.Data,
			}
			resource.Json, resource.Yaml, err = util.Marshal(configMap)
			if err != nil {
				klog.Errorf("Marshal %s %s error : %s", pkg.KubernetesResourceConfigMap, resource.Name, err)
				continue
			}

			if err = ss.FirstOrUpdate(resource); err != nil {
				klog.Errorln(err)
				continue
			}
		}

		token = configMapList.Continue
		if govalidator.IsNull(token) {
			break
		}
	}

	return nil
}

func (ss *saveService) CronJob(namespace string) error {
	// source CronJob client
	scjc := ss.skc.BatchV1beta1().CronJobs(namespace)

	// 处理 continue 参数
	token := pkg.DefaultToken
	var options metav1.ListOptions
	for {
		if token != pkg.DefaultToken {
			options = metav1.ListOptions{
				Continue: token,
			}
		} else {
			options = metav1.ListOptions{}
		}
		cronJobList, err := scjc.List(options)
		if err != nil {
			return err
		}

		for _, item := range cronJobList.Items {
			var resource model.Resource
			resource.Name = item.ObjectMeta.Name
			resource.Namespace = item.ObjectMeta.Namespace
			resource.Kind = pkg.KubernetesResourceCronJob
			resource.UUID = uuid.New().String()

			cronJob := &batchv1beta1.CronJob{
				TypeMeta: metav1.TypeMeta{
					Kind:       pkg.KubernetesResourceCronJob,
					APIVersion: pkg.KubernetesResourceCronJobAPIVersion,
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
			resource.Json, resource.Yaml, err = util.Marshal(cronJob)
			if err != nil {
				klog.Errorf("Marshal %s %s error : %s", pkg.KubernetesResourceCronJob, resource.Name, err)
				continue
			}

			if err = ss.FirstOrUpdate(resource); err != nil {
				klog.Errorln(err)
				continue
			}
		}

		token = cronJobList.Continue
		if govalidator.IsNull(token) {
			break
		}
	}

	return nil
}

func (ss *saveService) Namespace() error {
	// source namespace client
	snc := ss.skc.CoreV1().Namespaces()

	// 处理 continue 参数
	token := pkg.DefaultToken
	var options metav1.ListOptions
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
			var resource model.Resource
			resource.Name = item.ObjectMeta.Name
			resource.Kind = pkg.KubernetesResourceNamespace
			resource.UUID = uuid.New().String()
			resource.CreateTime = time.Now()
			resource.UpdateTime = time.Now()

			namespace := &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:   item.ObjectMeta.Name,
					Labels: item.ObjectMeta.Labels,
				},
			}
			resource.Json, resource.Yaml, err = util.Marshal(namespace)
			if err != nil {
				klog.Errorf("Marshal %s %s error : %s", pkg.KubernetesResourceNamespace, resource.Name, err)
				continue
			}

			if err = ss.FirstOrUpdate(resource); err != nil {
				klog.Errorln(err)
				continue
			}
		}

		token = namespaceList.Continue
		if govalidator.IsNull(token) {
			break
		}
	}

	return nil
}

func (ss *saveService) Secret(namespace string) error {
	// source Secret client
	ssc := ss.skc.CoreV1().Secrets(namespace)

	// 处理 continue 参数
	token := pkg.DefaultToken
	var options metav1.ListOptions
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
			var resource model.Resource
			resource.Name = item.ObjectMeta.Name
			resource.Namespace = item.ObjectMeta.Namespace
			resource.Kind = pkg.KubernetesResourceSecret
			resource.UUID = uuid.New().String()

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

			resource.Json, resource.Yaml, err = util.Marshal(secret)
			if err != nil {
				klog.Errorf("Marshal %s %s error : %s", pkg.KubernetesResourceSecret, resource.Name, err)
				continue
			}
			if err = ss.FirstOrUpdate(resource); err != nil {
				klog.Errorln(err)
				continue
			}
		}

		token = secretList.Continue
		if govalidator.IsNull(token) {
			break
		}
	}

	return nil
}

func (ss *saveService) FirstOrUpdate(r model.Resource) error {
	var t model.Resource

	err := ss.db.Model(&model.Resource{}).
		Where("namespace = ? AND kind = ? AND name = ?", r.Namespace, r.Kind, r.Name).
		First(&t).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound { // 增加
			r.IsCurrentUpdate = true
			if err := ss.db.Model(&model.Resource{}).Create(&r).Error; err != nil {
				return err
			}
		} else {
			return err
		}
	}

	if err := ss.db.Model(&model.Resource{}).
		Where("namespace = ? AND kind = ? AND name = ?", r.Namespace, r.Kind, r.Name).
		Update(map[string]interface{}{
			"json":              r.Json,
			"yaml":              r.Yaml,
			"is_current_update": true,
			"description":       r.Description,
			"update_time":       r.UpdateTime,
		}).Error; err != nil {
		return err
	}

	return nil
}

func (ss *saveService) afterSave() error {
	// 删除所有非当前更新的资源，也就是已经被删除的资源
	if err := ss.db.Model(&model.Resource{}).
		Where("is_current_update = ?", false).Delete(&model.Resource{}).Error; err != nil {
		return err
	}

	// 重置 is_current_update
	if err := ss.db.Model(&model.Resource{}).Update(map[string]interface{}{
		"is_current_update": false,
	}).Error; err != nil {
		return err
	}

	return nil
}

func (ss *saveService) Execute(nis mapset.Set, ris mapset.Set) {
	namespaces := util.Convert2Strings(nis.ToSlice())
	var err error
	for _, namespace := range namespaces {
		if ris.Contains(pkg.KubernetesResourceDeployment) {
			err = ss.Deployment(namespace)
			klog.Errorln(err)
		}
		if ris.Contains(pkg.KubernetesResourceService) {
			err = ss.Service(namespace)
			klog.Errorln(err)
		}
		if ris.Contains(pkg.KubernetesResourceIngress) {
			err = ss.Ingress(namespace)
			klog.Errorln(err)
		}
		if ris.Contains(pkg.KubernetesResourceConfigMap) {
			err = ss.ConfigMap(namespace)
			klog.Errorln(err)
		}
		if ris.Contains(pkg.KubernetesResourceCronJob) {
			err = ss.CronJob(namespace)
			klog.Errorln(err)
		}
		if ris.Contains(pkg.KubernetesResourceSecret) {
			err = ss.Secret(namespace)
			klog.Errorln(err)
		}
	}

	return
}
