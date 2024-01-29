## mysql 库与表创建

本项目需要依赖 mysql 来存储。以下提供创建 mysql 库与表的介绍与参考。

1. 必须先准备一个 mysql 类型的服务，并且集群网络可达，并在 helm value.yaml 中设置


```bash
[root@VM-0-16-centos yaml]# kubectl exec -it mydbconfig-controller-f484d984-wmr7q -- sh
Defaulted container "mysqltest" out of: mysqltest, mydbconfig
# mysql -uroot -p123456
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 121
Server version: 10.5.22-MariaDB-1:10.5.22+maria~ubu2004 mariadb.org binary distribution

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.
MariaDB [(none)]> drop database resources;
Query OK, 3 rows affected (0.006 sec)

MariaDB [(none)]> show databases;
+--------------------+
| Database           |
+--------------------+
| information_schema |
| mysql              |
| performance_schema |
+--------------------+
3 rows in set (0.000 sec)

MariaDB [(none)]> use resources;
Database changed
MariaDB [resources]>
MariaDB [resources]> DROP TABLE IF EXISTS `resources`;
Query OK, 0 rows affected, 1 warning (0.000 sec)

MariaDB [resources]> CREATE TABLE `resources`  (
    ->   `id` bigint(11) UNSIGNED NOT NULL AUTO_INCREMENT,
    ->   `namespace` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NULL DEFAULT NULL,
    ->   `name` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
    ->   `cluster` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
    ->   `group` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
    ->   `version` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
    ->   `resource` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
    ->   `kind` varchar(60) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
    ->   `resource_version` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NULL DEFAULT NULL,
    ->   `owner` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NULL DEFAULT NULL COMMENT 'owner uid',
    ->   `uid` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
    ->   `hash` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NULL DEFAULT NULL,
    ->   `object` json NULL,
    ->   `create_at` timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ->   `delete_at` timestamp(0) NULL DEFAULT NULL,
    ->   `update_at` timestamp(0) NULL DEFAULT NULL,
    ->   PRIMARY KEY (`id`) USING BTREE,
    ->   UNIQUE INDEX `uid`(`uid`) USING BTREE
    -> ) ENGINE = MyISAM AUTO_INCREMENT = 0 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_bin ROW_FORMAT = Dynamic;
Query OK, 0 rows affected (0.004 sec)

MariaDB [resources]> DROP TABLE IF EXISTS `clusters`;
Query OK, 0 rows affected, 1 warning (0.000 sec)

MariaDB [resources]> CREATE TABLE `clusters`  (
    ->   `id` bigint(11) UNSIGNED NOT NULL AUTO_INCREMENT,
    ->   `name` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
    ->   `isMaster` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
    ->   `status` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
    ->   `create_at` timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ->   PRIMARY KEY (`id`) USING BTREE,
    ->   UNIQUE INDEX `name`(`name`) USING BTREE
    -> ) ENGINE = MyISAM AUTO_INCREMENT = 0 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_bin ROW_FORMAT = Dynamic;
Query OK, 0 rows affected (0.004 sec)

MariaDB [resources]>
```


```bash
[root@VM-0-16-centos helm]# kubectl logs multi-cluster-operator-6c6fc9dcc9-xs79f
F0129 10:02:52.587503       1 mysql_options.go:144] Failed to apply migrations: Dirty database version 1. Fix and force version.
[root@VM-0-16-centos helm]# kubectl get pods

使用

MariaDB [resources]> UPDATE schema_migrations SET dirty = 0;
Query OK, 1 row affected (0.001 sec)
Rows matched: 1  Changed: 1  Warnings: 0

```