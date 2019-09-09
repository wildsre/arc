create table arc_user (
    id int(15) not null auto_increment,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
    deleted_at timestamp NULL DEFAULT NULL,
    deleted tinyint(1) default 0,
    name char(127),
    token char(127),
    role char(31) comment 'superadmin, admin, user',
    default_space char(127),
    user_group text comment 'json: user group and auth { "group1":"",}',
    space text comment 'json: spacename and aut h{"default":"760"}, 7: read, write, delete; 6: read, write; 4: read; 2: write; 1: delete',
    PRIMARY KEY (`id`),
    UNIQUE KEY `username` (`name`)
) DEFAULT CHARSET=utf8mb4;


create table arc_usergroup (
    id int(15) not null auto_increment,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
    deleted_at timestamp NULL DEFAULT NULL,
    deleted tinyint(1) default 0,
    `name` char(127),
    `member` char(127),
    `auth` char(3) comment '2: write add or remove user from group; 1: delete group',
    `desc` text,
    PRIMARY KEY (`id`),
    UNIQUE KEY `groupname` (`name`)
) DEFAULT CHARSET=utf8mb4;
    

create table arc_data (
    id int(15) not null auto_increment,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
    deleted_at timestamp NULL DEFAULT NULL,
    deleted tinyint(1) default 0,
    space char(255),
    `app` char(255),
    `resource` char(255),
    `item` char(255),
    `value` text,
    `value_tag` text,
    PRIMARY KEY (`id`)
) DEFAULT CHARSET=utf8mb4;
