DROP TABLE IF EXISTS `resources`;
CREATE TABLE `resources`  (
                              `id` bigint(11) UNSIGNED NOT NULL AUTO_INCREMENT,
                              `namespace` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NULL DEFAULT NULL,
                              `name` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
                              `cluster` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
                              `group` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
                              `version` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
                              `resource` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
                              `kind` varchar(60) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
                              `resource_version` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NULL DEFAULT NULL,
                              `owner` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NULL DEFAULT NULL COMMENT 'owner uid',
                              `uid` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
                              `hash` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NULL DEFAULT NULL,
                              `object` json NULL,
                              `create_at` timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP,
                              `delete_at` timestamp(0) NULL DEFAULT NULL,
                              `update_at` timestamp(0) NULL DEFAULT NULL,
                              PRIMARY KEY (`id`) USING BTREE,
                              UNIQUE INDEX `uid`(`uid`) USING BTREE
) ENGINE = MyISAM AUTO_INCREMENT = 0 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_bin ROW_FORMAT = Dynamic;


DROP TABLE IF EXISTS `clusters`;
CREATE TABLE `clusters`  (
                             `id` bigint(11) UNSIGNED NOT NULL AUTO_INCREMENT,
                             `name` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
                             `isMaster` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
                             `status` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
                             `create_at` timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP,
                             PRIMARY KEY (`id`) USING BTREE,
                             UNIQUE INDEX `name`(`name`) USING BTREE
) ENGINE = MyISAM AUTO_INCREMENT = 0 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_bin ROW_FORMAT = Dynamic;


