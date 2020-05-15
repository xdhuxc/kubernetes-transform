USE `kubernetes`;

delete from `k8s_resource`;

ALTER TABLE `k8s_resource` ADD COLUMN `uuid`      varchar(256) NOT NULL COMMENT 'uuid';

ALTER TABLE `k8s_resource` ADD COLUMN `is_current_update` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否为此次更新';

ALTER TABLE `k8s_resource` DROP COLUMN `version`;

ALTER TABLE `k8s_resource` DROP COLUMN `is_last`;




