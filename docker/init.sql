insert into product (product_id, name, price, create_date, update_date)
values (uuid_generate_v4(), 'Harry Potter and the goblet of fire', 20, now(), now());

insert into product (product_id, name, price, create_date, update_date)
values (uuid_generate_v4(), 'Harry potter and the chamber of secrets', 10, now(), now());

insert into account (account_id, amount, update_date)
values (uuid_generate_v4(), 10000, now());