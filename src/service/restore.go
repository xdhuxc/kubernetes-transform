package service

import (
	"fmt"
	"strings"

	mapset "github.com/deckarep/golang-set"
	"github.com/jinzhu/gorm"
	appv1 "k8s.io/api/apps/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	"github.com/xdhuxc/kubernetes-transform/src/config"
	"github.com/xdhuxc/kubernetes-transform/src/model"
	"github.com/xdhuxc/kubernetes-transform/src/pkg"
	"github.com/xdhuxc/kubernetes-transform/src/util"
)

type restoreService struct {
	cnf config.Config
	tc  model.Cluster
	tkc *kubernetes.Clientset // target kubernetes cluster client
	db  *gorm.DB
}

func newRestoreService(cnf config.Config, tkc *kubernetes.Clientset, db *gorm.DB) *restoreService {
	return &restoreService{
		cnf: cnf,
		db:  db,
		tc:  cnf.Target,
		tkc: tkc,
	}
}

func (rs *restoreService) Restore() error {
	var resources []model.Resource
	if err := rs.db.Model(&model.Resource{}).Find(&resources).Error; err != nil {
		return err
	}
	_, nss, err := Namespaces(rs.tkc)
	if err != nil {
		return err
	}

	namespaces := mapset.NewSetFromSlice(util.Convert2Interfaces(nss))
	// Namespace Request Set
	nrs := mapset.NewSetFromSlice(util.Convert2Interfaces(rs.cnf.Namespace.Namespaces))
	kinds := mapset.NewSetFromSlice(util.Convert2Interfaces(rs.cnf.Resource.Kinds))
	// Resource Request Set
	rrs := mapset.NewSetFromSlice(util.Convert2Interfaces(rs.cnf.Resource.Resources))

	if rs.cnf.Namespace.Action == pkg.KubernetesResourceActionInclude {
		// Namespace Inclusion Set
		nis := namespaces.Intersect(nrs)
		if nis.Cardinality() > 0 {
			// nss := util.Convert2Strings(nis.ToSlice())
			if rs.cnf.Resource.Action == pkg.KubernetesResourceActionInclude {
				// Resource Inclusion Set
				ris := kinds.Intersect(rrs)
				if ris.Cardinality() > 0 {
					rs.Execute(nis, ris, resources)
				} else {
					return fmt.Errorf("there is no intersection between the requested resources and resources that has been coded currently")
				}
			} else if rs.cnf.Resource.Action == pkg.KubernetesResourceActionExclude {
				// Resource Inclusion Set
				ris := kinds.Difference(rrs)
				if ris.Cardinality() > 0 {
					rs.Execute(nis, ris, resources)
				} else {
					return fmt.Errorf("the difference between the requested resources and the qctual resources is an empty set")
				}
			} else {
				return fmt.Errorf("the action of resource request is invalid")
			}
		} else {
			return fmt.Errorf("there is no intersection between the requested namespaces and the actual namespaces")
		}
	} else if rs.cnf.Namespace.Action == pkg.KubernetesResourceActionExclude {
		// Namespace Inclusion Set
		nis := namespaces.Difference(nrs)
		if nis.Cardinality() > 0 {
			// nss := util.Convert2Strings(nis.ToSlice())
			if rs.cnf.Resource.Action == pkg.KubernetesResourceActionInclude {
				// Resource Inclusion Set
				ris := kinds.Intersect(rrs)
				if ris.Cardinality() > 0 {
					rs.Execute(nis, ris, resources)
				} else {
					return fmt.Errorf("there is no intersection between the requested resources and resources that has been coded currently")
				}
			} else if rs.cnf.Resource.Action == pkg.KubernetesResourceActionExclude {
				// Resource Inclusion Set
				ris := kinds.Difference(rrs)
				if ris.Cardinality() > 0 {
					rs.Execute(nis, ris, resources)
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

func (rs *restoreService) Deployment(r model.Resource) error {
	// 从目标集群获取此 deployment
	getOptions := metav1.GetOptions{}
	_, err := rs.tkc.AppsV1().Deployments(r.Namespace).Get(r.Name, getOptions)
	if err == nil {
		klog.Infof("the deployment %s has already existed in target cluster %s", r.Name, rs.tc.Name)
		return nil
	}

	var deployment appv1.Deployment
	reader := strings.NewReader(r.Yaml)
	decoder := yaml.NewYAMLOrJSONDecoder(reader, 4096)
	err = decoder.Decode(&deployment)
	if err != nil {
		return err
	}

	// 创建 deployment 到目标集群
	_, err = rs.tkc.AppsV1().Deployments(r.Namespace).Create(&deployment)
	if err != nil {
		return err
	}

	return nil
}

func (rs *restoreService) Service(r model.Resource) error {
	// 从目标集群获取此 service
	getOptions := metav1.GetOptions{}
	_, err := rs.tkc.CoreV1().Services(r.Namespace).Get(r.Name, getOptions)
	if err == nil {
		klog.Infof("the service %s has already existed in target cluster %s", r.Name, rs.tc.Name)
		return nil
	}

	var service v1.Service
	reader := strings.NewReader(r.Yaml)
	decoder := yaml.NewYAMLOrJSONDecoder(reader, 4096)
	err = decoder.Decode(&service)
	if err != nil {
		return err
	}

	// 创建 service 到目标集群
	_, err = rs.tkc.CoreV1().Services(r.Namespace).Create(&service)
	if err != nil {
		return err
	}

	return nil
}

func (rs *restoreService) Ingress(r model.Resource) error {
	// 从目标集群获取此 ingress
	getOptions := metav1.GetOptions{}
	_, err := rs.tkc.ExtensionsV1beta1().Ingresses(r.Namespace).Get(r.Name, getOptions)
	if err == nil {
		klog.Infof("the ingress %s has already existed in target cluster %s", r.Name, rs.tc.Name)
		return nil
	}

	var ingress v1beta1.Ingress
	reader := strings.NewReader(r.Yaml)
	decoder := yaml.NewYAMLOrJSONDecoder(reader, 4096)
	err = decoder.Decode(&ingress)
	if err != nil {
		return err
	}

	// 创建 ingress 到目标集群
	_, err = rs.tkc.ExtensionsV1beta1().Ingresses(r.Namespace).Create(&ingress)
	if err != nil {
		return err
	}

	return nil
}

func (rs *restoreService) ConfigMap(r model.Resource) error {
	// 从目标集群获取此 configMap
	getOptions := metav1.GetOptions{}
	_, err := rs.tkc.CoreV1().ConfigMaps(r.Namespace).Get(r.Name, getOptions)
	if err == nil {
		klog.Infof("the configMap %s has already existed in target cluster %s", r.Name, rs.tc.Name)
		return nil
	}

	var configMap v1.ConfigMap
	reader := strings.NewReader(r.Yaml)
	decoder := yaml.NewYAMLOrJSONDecoder(reader, 4096)
	err = decoder.Decode(&configMap)
	if err != nil {
		return err
	}

	// 创建 configMap 到目标集群
	_, err = rs.tkc.CoreV1().ConfigMaps(r.Namespace).Create(&configMap)
	if err != nil {
		return err
	}

	return nil
}

func (rs *restoreService) CronJob(r model.Resource) error {
	// 从目标集群获取此 cronJob
	getOptions := metav1.GetOptions{}
	_, err := rs.tkc.BatchV1beta1().CronJobs(r.Namespace).Get(r.Name, getOptions)
	if err == nil {
		klog.Infof("the cronJob %s has already existed in target cluster %s", r.Name, rs.tc.Name)
		return nil
	}

	var cronJob batchv1beta1.CronJob
	reader := strings.NewReader(r.Yaml)
	decoder := yaml.NewYAMLOrJSONDecoder(reader, 4096)
	err = decoder.Decode(&cronJob)
	if err != nil {
		return err
	}

	// 创建 cronJob 到目标集群
	_, err = rs.tkc.BatchV1beta1().CronJobs(r.Namespace).Create(&cronJob)
	if err != nil {
		return err
	}

	return nil
}

func (rs *restoreService) Namespace(r model.Resource) error {
	// 从目标集群获取此 namespace
	getOptions := metav1.GetOptions{}
	_, err := rs.tkc.CoreV1().Namespaces().Get(r.Name, getOptions)
	if err == nil {
		klog.Infof("the namespace %s has already existed in target cluster %s", r.Name, rs.tc.Name)
		return nil
	}

	var namespace v1.Namespace
	reader := strings.NewReader(r.Yaml)
	decoder := yaml.NewYAMLOrJSONDecoder(reader, 4096)
	err = decoder.Decode(&namespace)
	if err != nil {
		return err
	}

	// 创建 namespace 到目标集群
	_, err = rs.tkc.CoreV1().Namespaces().Create(&namespace)
	if err != nil {
		return err
	}

	return nil
}

func (rs *restoreService) Secret(r model.Resource) error {
	// 从目标集群获取此 secret
	getOptions := metav1.GetOptions{}
	_, err := rs.tkc.CoreV1().Secrets(r.Namespace).Get(r.Name, getOptions)
	if err == nil {
		klog.Infof("the secret %s has already existed in target cluster %s", r.Name, rs.tc.Name)
		return nil
	}

	var secret v1.Secret
	reader := strings.NewReader(r.Yaml)
	decoder := yaml.NewYAMLOrJSONDecoder(reader, 4096)
	err = decoder.Decode(&secret)
	if err != nil {
		return err
	}

	// 创建 secret 到目标集群
	_, err = rs.tkc.CoreV1().Secrets(r.Namespace).Create(&secret)
	if err != nil {
		return err
	}

	return nil
}

func (rs *restoreService) Execute(nis mapset.Set, ris mapset.Set, resources []model.Resource) {
	var err error
	for _, resource := range resources {
		if nis.Contains(resource.Namespace) && ris.Contains(resource.Kind) {
			err = rs.Deployment(resource)
			klog.Errorln(err)
		}
		if nis.Contains(resource.Namespace) && ris.Contains(resource.Kind) {
			err = rs.Service(resource)
			klog.Errorln(err)
		}
		if nis.Contains(resource.Namespace) && ris.Contains(resource.Kind) {
			err = rs.Ingress(resource)
			klog.Errorln(err)
		}
		if nis.Contains(resource.Namespace) && ris.Contains(resource.Kind) {
			err = rs.ConfigMap(resource)
			klog.Errorln(err)
		}
		if nis.Contains(resource.Namespace) && ris.Contains(resource.Kind) {
			err = rs.CronJob(resource)
			klog.Errorln(err)
		}
		if nis.Contains(resource.Namespace) && ris.Contains(resource.Kind) {
			err = rs.Secret(resource)
			klog.Errorln(err)
		}
	}

	return
}
