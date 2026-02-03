DROP TABLE IF EXISTS `t_admin_user`;
CREATE TABLE `t_admin_user`
(
    `id`            BIGINT       NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '管理员ID',
    `username`      VARCHAR(50)  NOT NULL DEFAULT '' COMMENT '管理员用户名',
    `password`      VARCHAR(255) NOT NULL DEFAULT '' COMMENT '加密后的密码（bcrypt）',
    `real_name`     VARCHAR(100) NOT NULL DEFAULT '' COMMENT '真实姓名',
    `email`         VARCHAR(100) DEFAULT NULL COMMENT '邮箱',
    `phone`         VARCHAR(20)  DEFAULT NULL COMMENT '手机号',
    `role`          VARCHAR(50)  NOT NULL DEFAULT 'admin' COMMENT '角色：super_admin-超级管理员 admin-普通管理员',
    `status`        TINYINT      NOT NULL DEFAULT 1 COMMENT '状态：0-禁用 1-正常',
    `last_login_at` TIMESTAMP NULL COMMENT '最后登录时间',
    `last_login_ip` VARCHAR(50)  NOT NULL DEFAULT '' COMMENT '最后登录IP',
    `created_at`    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY `uk_username` (`username`),
    UNIQUE KEY `uk_email` (`email`),
    UNIQUE KEY `uk_phone` (`phone`),
    INDEX `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理员表';
