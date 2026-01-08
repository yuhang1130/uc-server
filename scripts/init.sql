CREATE TABLE `user` (
  `id` BIGINT UNSIGNED NOT NULL PRIMARY KEY comment "主键ID",
  `username` VARCHAR(50) NOT NULL comment "用户名（如：admin）",
  `password_hash` VARCHAR(255) NOT NULL comment "密码哈希值",
  `email` VARCHAR(100) NOT NULL comment "邮箱地址",
  `role` VARCHAR(50) NOT NULL  DEFAULT 'user' COMMENT "全局角色：super_admin,user",
  `status` TINYINT DEFAULT 1 COMMENT '状态：1-正常, 0-停用',
  `created_at` BIGINT NOT NULL comment "创建时间",
  `updated_at` BIGINT NOT NULL comment "更新时间",
  `deleted_at` BIGINT DEFAULT NULL comment "删除时间",
  UNIQUE KEY `idx_username` (`username`, `deleted_at`),
  UNIQUE KEY `idx_email` (`email`, `deleted_at`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 comment = "用户表（系统账号）";


/* user中的role 它只用于标识 全局身份（即：不依赖任何租户就能确定的角色）
  super_admin：系统管理员（可管理所有租户）
  user：普通用户（必须结合 tenant_user 才能知道他在某个租户中的具体角色）
*/

-- 初始化系统管理员账号
-- 默认密码: admin@123 (请在首次登录后修改)
-- 密码使用 bcrypt 加密，成本因子为 10
INSERT INTO `user` (
  `id`,
  `username`,
  `password_hash`,
  `email`,
  `role`,
  `status`,
  `created_at`,
  `updated_at`,
  `deleted_at`
)
VALUES (
  1,
  'admin',
  '$2a$10$QEQTPnIOjxMz9mi5h8DJn.vHQyTgMR/WEhKe12fKhbyhQMWUc6RAm',  -- 密码: admin@123
  'hongdou_hyh@163.com',
  'super_admin',
  1,
  UNIX_TIMESTAMP(NOW()) * 1000,
  UNIX_TIMESTAMP(NOW()) * 1000,
  NULL
);

CREATE TABLE `tenant` (
  `id` BIGINT UNSIGNED NOT NULL PRIMARY KEY comment "主键ID",
  `tenant_name` VARCHAR(100) NOT NULL COMMENT '租户名称（如: ABC公司）',
  `status` TINYINT DEFAULT 1 COMMENT '状态: 1-正常, 0-停用',
  `created_at` BIGINT NOT NULL comment "创建时间",
  `updated_at` BIGINT NOT NULL comment "更新时间",
  `deleted_at` BIGINT DEFAULT NULL comment "删除时间",
  UNIQUE KEY `idx_tenant_name` (`tenant_name`, `deleted_at`) -- 防止同名租户
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='租户表（租户是企业）';

CREATE TABLE `tenant_user` (
  `tenant_id` BIGINT UNSIGNED NOT NULL COMMENT '租户ID（关联tenant表）',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID（关联user表）',
  `role` VARCHAR(50) NOT NULL COMMENT '角色: tenant_admin（租户管理员）, user（普通员工）',
  `status` TINYINT DEFAULT 1 COMMENT '状态: 1-正常, 0-停用',
  `created_at` BIGINT NOT NULL comment "创建时间",
  `updated_at` BIGINT NOT NULL comment "更新时间",
  `deleted_at` BIGINT DEFAULT NULL comment "删除时间",
  `last_login_at` BIGINT NULL comment "最后一次登录时间",
  PRIMARY KEY (`tenant_id`, `user_id`), -- 一个用户在同一个租户只能有1个角色
  KEY `idx_user` (`user_id`),
  KEY `idx_role` (`role`),
  CONSTRAINT `fk_tenant_user_tenant` FOREIGN KEY (`tenant_id`) REFERENCES `tenant`(`tenant_id`),
  CONSTRAINT `fk_tenant_user_user` FOREIGN KEY (`user_id`) REFERENCES `user`(`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='租户与用户关联表（表示用户属于哪个租户）';

CREATE TABLE `employee` (
  `id` BIGINT UNSIGNED NOT NULL PRIMARY KEY comment "主键ID",
  `tenant_id` BIGINT UNSIGNED NOT NULL COMMENT '租户ID（关联tenant表）',
  `user_id` BIGINT UNSIGNED NULL COMMENT '用户ID（关联user表）',
  `name` VARCHAR(50) NOT NULL COMMENT '姓名',
  `position` VARCHAR(100) COMMENT '职位',
  `department` VARCHAR(100) COMMENT '部门',
  `email` VARCHAR(100) COMMENT '邮箱（租户内唯一，非全局唯一）',
  `status` TINYINT DEFAULT 1 COMMENT '状态: 1-在职, 0-离职',
  `created_at` BIGINT NOT NULL comment "创建时间",
  `updated_at` BIGINT NOT NULL comment "更新时间",
  `deleted_at` BIGINT DEFAULT NULL comment "删除时间",
  KEY `idx_tenant` (`tenant_id`),
  KEY `idx_email` (`email`, `deleted_at`), -- 租户内邮箱唯一
  CONSTRAINT `fk_employee_tenant` FOREIGN KEY (`tenant_id`) REFERENCES `tenant`(`tenant_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='员工表（存储组织内部的员工信息）';