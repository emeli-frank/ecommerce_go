CREATE TABLE users
(
    id SERIAL,
    first_name VARCHAR(64) NOT NULL,
    last_name VARCHAR (64) NOT NULL,
    email VARCHAR (128) NOT NULL,
    password CHAR(60) NOT NULL,
    address_id int,

    PRIMARY KEY (id),
    UNIQUE (email),
    FOREIGN KEY (address_id)
        REFERENCES addresses (id)
        ON DELETE CASCADE
);

CREATE TABLE addresses
(
    id SERIAL,
    country VARCHAR(32) NOT NULL,
    state VARCHAR (32) NOT NULL,
    city VARCHAR (32) NOT NULL,
    postal_code VARCHAR (8) NOT NULL,
    address VARCHAR (64) NOT NULL,

    PRIMARY KEY (id)
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

    PRIMARY KEY (id),
    FOREIGN KEY (category_id)
        REFERENCES product_categories (id)
        ON DELETE CASCADE
);

CREATE TABLE credit_cards
(
    id SERIAL,
    customer_id int NOT NULL,
    name varchar (64) NOT NULL,
    number varchar (20) NOT NULL,
    cvc char(3) NOT NULL,
    expiry_date timestamptz,

    PRIMARY KEY(id),
    FOREIGN KEY (customer_id)
        REFERENCES users (id)
        ON DELETE CASCADE
);
