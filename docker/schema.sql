create table products
(
    product_id varchar(512) not null,
    name varchar(128),
    price numeric,
    create_date date,
    update_date date
);

create table product_orders
(
    product_order_id varchar(512) not null,
    product_id varchar(512) not null,
    account_id varchar(512) not null,
    create_date date,
    request_id varchar(512) not null
);

create table accounts
(
    account_id varchar(512) not null,
    amount numeric,
    update_date date
);

create table transactions
(
    transaction_id varchar(512) not null,
    product_id varchar(512) not null,
    amount numeric,
    create_date date,
    request_id varchar(512) not null,
    account_id varchar(512) not null
);

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";