CREATE TABLE users
(
    id SERIAL,
    first_name VARCHAR(64) NOT NULL,
    last_name VARCHAR (64) NOT NULL,
    email VARCHAR (128) NOT NULL,
    password CHAR(60) NOT NULL
    PRIMARY KEY (id),
    UNIQUE (email)
);

CREATE SEQUENCE role_id_seq;
CREATE TABLE roles
(
    id         smallint NOT NULL DEFAULT nextval('role_id_seq'),
    name       VARCHAR(32) NOT NULL,

    PRIMARY KEY (id)
);

CREATE TABLE role_user_map
(
    user_id INT NOT NULL,
    role_id smallint NOT NULL,

    UNIQUE (user_id, role_id),
    FOREIGN KEY (role_id)
        REFERENCES roles (id)
        ON DELETE CASCADE,
    FOREIGN KEY (user_id)
        REFERENCES users (id)
        ON DELETE CASCADE
);

CREATE TABLE product_categories
(
    id SERIAL,
    name VARCHAR(32) NOT NULL,

    PRIMARY KEY(id)
);

CREATE TABLE products
(
    id SERIAL,
    name varchar (32) NOT NULL,
    category_id int NOT NULL,
    price float NOT NULL,
    old_price float,
    rating smallint,
    description varchar(2048),
    quantity int,

    FOREIGN KEY (category_id)
        REFERENCES product_categories (id)
        ON DELETE CASCADE
);