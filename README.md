### kubernetes-transform

可以实现两个 kubernetes 集群之间资源的复制

程序的运行模式可以为：

* transform：将源集群中的资源创建到目标集群中，同名资源需要指定策略（默认为：skip）
* save：将源集群中的资源保存到数据库中，也是默认模式
* restore：将数据库中的源集群资源创建到目标集群中，同名资源需要指定策略（默认为：skip）

各模式可以使用接口调用，调用时可以更改同名资源策略，内部使用定时任务运行在 save 模式下。

同名资源创建策略为：

* skip：跳过，同名资源不创建，默认选项
* merge：合并同名资源
* update：更新同名资源

可以通过 --namespace.exclusions 参数指定排除的命名空间 
可以通过 --resource.exclusions 参数指定排除的资源

可以通过如下数据格式指定排除的命名空间或资源：
```markdown
{
    "name": "namespace"
    "action": "include" # or exclude
    "namespaces": ["kube-admin", "kube-system", "kube-public"]
}
```
或
```markdown
{
    "name": "resource",
    "action": "exclude" # or exclude
    "resources": ["Deployment", "Service"]
}
```
可以通过 exclusions 来指定排除的资源

使用 inclusions 来指定需要保存的资源，

在默认情况下，将会保存各个命名空间中的各种资源，提供这两个参数主要是为了只保存某一种或某几种资源，或者排除某一命名空间或某几个命名空间时使用

注意，一次请求中，只能指定 inclusions 或 exclusions

#### 标签的删除，替换和合并

删除标签的写法如下：
```markdown
{
    "action": "delete",
    "labels" : {
        "env": "prod",
        "group": "ADS"
    }
}
```
如果源资源含有这些标签，则目标资源将不会有这些标签

替换标签的写法如下：
```markdown
{
    "action": "update",
    "labels" : {
        "env": "prod",
        "group": "ADS"
    }
}
```
如果源资源有这些标签，则目标资源标签的值将会更新为这些标签中的同名标签的值，如果源资源没有该标签，则不予处理。

合并标签的写法如下：
```markdown
{
    "action": "merge",
    "labels" : {
        "env": "prod",
        "group": "ADS"
    }
}
```
目标资源中将会带有这些新标签，如果含有同名标签，则会覆盖旧值。

### 功能
kubernetes-transform 的功能：
* 将源集群中的资源创建到目标集群中，可以指定策略处理同名资源
* 将源集群中的资源数据保存到数据库中
* 将数据库中的集群资源创建到目标集群中，可以指定策略处理同名资源
* 支持原生 Kubernetes 资源
* 支持 API 调用
* 可以指定同名资源创建策略，包括：skip，merge，update
* 支持命名空间排除和资源排除
* 支持标签排除，替换和合并
* 支持 AWS EKS 和 华为云 CCE
* 支持配置文件配置和接口调整配置

### 注意事项

1、对于 NodeSelector 标签的处理

### 待完成的工作

1、代码结构优化

2、目前完成的资源：
* Deployment
* Service
* Ingress

未完成的资源：
* Namespace
* ConfigMap
* CronJob
* ServiceAccount
* PVC
* StatefulSet
* Secret
* HPA（nativeHPA，eventHPA，configHPA）

3、将 kubernetes 资源定时备份到数据库 

4、从数据库快速创建集群资源

### 注意事项

本项目尚未全部完成
