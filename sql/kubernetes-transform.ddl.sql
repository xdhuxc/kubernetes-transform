CREATE DATABASE `kubernetes`;

USE `kubernetes`;

# DDL
CREATE TABLE IF NOT EXISTS `k8s_cluster`
(
    `id`                int(11)      NOT NULL AUTO_INCREMENT COMMENT '自增 id',
    `uuid`      varchar(256) NOT NULL COMMENT 'uuid',
    `name`              varchar(256) NOT NULL COMMENT '集群名称',

    `address`      varchar(256)          DEFAULT '' COMMENT '集群地址',
    `token`      varchar(256)          DEFAULT '' COMMENT '集群 Token',
    `cloud`      varchar(256)          DEFAULT '' COMMENT '公有云服务商',
    `region`      varchar(256)          DEFAULT '' COMMENT '集群区域',

    `description`     varchar(256)           DEFAULT '' COMMENT '说明信息',
    `create_time`       datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time`       datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8;

CREATE TABLE IF NOT EXISTS `k8s_resource`
(
    `id`                int(11)      NOT NULL AUTO_INCREMENT COMMENT '自增 id',
    `uuid`      varchar(256) NOT NULL COMMENT 'uuid',
    `name`              varchar(256) NOT NULL COMMENT '名称',

    `kind`      varchar(256) NOT NULL COMMENT '资源',
    `namespace`   varchar(256)          DEFAULT '' COMMENT '命名空间',
    `json` text      COMMENT '资源定义的 JSON 格式',
    `yaml` text      COMMENT '资源定义的 YAML 格式',
    `is_current_update` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否为此次更新',

    `description`     varchar(256)           DEFAULT '' COMMENT '说明信息',
    `create_time`       datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time`       datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    PRIMARY KEY (`id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8;











