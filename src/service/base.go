package service

import (
	"github.com/asaskevich/govalidator"
	"github.com/jinzhu/gorm"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/xdhuxc/kubernetes-transform/src/client"
	"github.com/xdhuxc/kubernetes-transform/src/config"
	"github.com/xdhuxc/kubernetes-transform/src/pkg"
)

type BaseService struct {
	HealthCheckService *healthCheckService
	TransformService   *transformService
	SaveService        *saveService
	RestoreService     *restoreService
}

func NewBaseService(cnf config.Config, db *gorm.DB) (*BaseService, error) {
	skc, err := client.NewKubernetesClusterClient(cnf.Source.Name, cnf.Source.Address, cnf.Source.Token)
	if err != nil {
		return nil, err
	}
	tkc, err := client.NewKubernetesClusterClient(cnf.Target.Name, cnf.Target.Address, cnf.Target.Token)
	if err != nil {
		return nil, err
	}

	return &BaseService{
		HealthCheckService: newHealthCheckService(db),
		TransformService:   newTransformService(cnf, skc, tkc),
		SaveService:        newSaveService(cnf, skc, db),
		RestoreService:     newRestoreService(cnf, tkc, db),
	}, nil
}

func Namespaces(skc *kubernetes.Clientset) ([]v1.Namespace, []string, error) {
	// source namespace client
	snc := skc.CoreV1().Namespaces()

	var namespaces []v1.Namespace
	var nss []string
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
			return nil, nil, err
		}
		namespaces = append(namespaces, namespaceList.Items...)

		token = namespaceList.Continue
		if govalidator.IsNull(token) {
			break
		}
	}

	for _, item := range namespaces {
		nss = append(nss, item.ObjectMeta.Name)
	}

	return namespaces, nss, nil
}
